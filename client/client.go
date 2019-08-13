package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	kuhnuri "github.com/kuhnuri/go-worker"
	"github.com/kuhnuri/kuhnuri-cli/config"
	"github.com/kuhnuri/kuhnuri-cli/models"
	"github.com/kuhnuri/kuhnuri-cli/spinner"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Client struct {
	base      string
	transtype string
	input     *url.URL
	output    *url.URL
	spinner   *spinner.Spinner
}

func New(conf config.Config, transtype string, input *url.URL, output *url.URL) (*Client, error) {
	if _, ok := conf["api"]; !ok {
		return nil, fmt.Errorf("API base URL not configured\n")
	}
	return &Client{
		base:      conf["api"],
		transtype: transtype,
		input:     input,
		output:    output,
	}, nil
}

func (c *Client) Execute() error {
	c.spinner = spinner.New("Running")
	defer c.spinner.Stop()
	var in *url.URL
	if isLocal(c.input) {
		var err error
		in, err = c.upload()
		if err != nil {
			return err
		}
	} else {
		in = c.input
	}
	c.spinner.Message("Submitting")
	create := models.NewCreate([]string{c.transtype}, in, c.output)
	job, err := c.doCreate(create)
	if err != nil {
		return fmt.Errorf("Failed to submit conversion: %v", err)
	}
	err = c.await(job)
	if err != nil {
		return err
	}
	if isExternal(c.output) {
		c.spinner.Message(fmt.Sprintf("Generated %s", job.Output))
	} else {
		err := c.download(job.Id)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Client) download(id string) error {
	c.spinner.Message(fmt.Sprintf("Downloading"))
	dst, err := c.doDownload(id)
	if err != nil {
		return fmt.Errorf("Failed to download results: %v", err)
	}
	c.spinner.Message(fmt.Sprintf("Unpackaging %s", dst))
	file := toPath(c.output)
	err = kuhnuri.MkDirs(file)
	if err != nil {
		return err
	}
	err = kuhnuri.Unzip(dst, file)
	if err != nil {
		return fmt.Errorf("Failed to unpackage %s: %v", dst, err)
	}
	return nil
}

func (c *Client) upload() (*url.URL, error) {
	file := toPath(c.input)
	stat, err := os.Stat(file)
	if err != nil {
		return nil, err
	}
	var dir string
	var start string
	if stat.IsDir() {
		dir = file
		filepath.Walk(file, func(path string, info os.FileInfo, err error) error {
			if filepath.Ext(path) == ".ditamap" && start == "" {
				start, err = filepath.Rel(file, path)
				if err != nil {
					return fmt.Errorf("Failed to build relative path for %s", path)
				}
				return filepath.SkipDir
			}
			return nil
		})
	} else {
		file := toPath(c.input)
		dir = filepath.Dir(file)
		start, err = filepath.Rel(dir, file)
		if err != nil {
			return nil, fmt.Errorf("Failed to build relative path for %s", c.input)
		}
	}
	c.spinner.Message(fmt.Sprintf("Packaging"))
	zip, err := createPackage(dir)
	if err != nil {
		return nil, fmt.Errorf("Failed to package source: %v", err)
	}
	c.spinner.Message("Uploading")
	upload, err := c.getUpload()
	if err != nil {
		return nil, fmt.Errorf("Failed to retrieve upload URL: %v", err)
	}
	err = doUpload(zip, upload.Upload)
	if err != nil {
		return nil, fmt.Errorf("Failed to upload source: %v", err)
	}

	in, err := url.Parse(fmt.Sprintf("jar:%s!/%s", upload.Url, filepath.ToSlash(start)))
	if err != nil {
		return nil, fmt.Errorf("Failed to construct URL: %v", err)
	}
	return in, nil
}

// Check if input argument is non-local URI resource
func isExternal(in *url.URL) bool {
	if !in.IsAbs() {
		return false
	}
	if in.Scheme == "file" {
		return false
	}
	return true
}

// Check if input argument is local URI resource
func isLocal(in *url.URL) bool {
	if !in.IsAbs() {
		return false
	}
	if in.Scheme != "file" {
		return false
	}
	_, err := os.Stat(toPath(in))
	if os.IsNotExist(err) {
		return false
	}
	return true
}

func toPath(in *url.URL) string {
	return filepath.FromSlash(in.Path)
}

func (c *Client) doDownload(in string) (string, error) {
	tmp, err := ioutil.TempFile("", "*.zip")
	if err != nil {
		return "", err
	}
	out, err := os.Create(tmp.Name())
	defer out.Close()
	if err != nil {
		return "", err
	}
	api := c.api("download", in)
	resp, err := http.Get(api)
	defer resp.Body.Close()
	if err != nil {
		return "", err
	}
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return "", err
	}
	return tmp.Name(), nil
}

func (c *Client) await(created *models.Job) error {
	id := created.Id

	//c.spinner.Message(fmt.Sprintf("Converting %s", id))
	ticks := time.Tick(5 * time.Second)
	for range ticks {
		job, err := c.getJob(id)
		if err != nil {
			return fmt.Errorf("Failed to retrieve state: %v", err)
		}
		switch job.Status {
		case "queue":
			c.spinner.Message(fmt.Sprintf("Queuing %s", id))
		case "process":
			c.spinner.Message(fmt.Sprintf("Converting %s", id))
			break
		case "done":
			//c.spinner.Message(fmt.Sprintf("Done %s", id))
			return nil
		case "error":
			//c.spinner.Message(fmt.Sprintf("Failed %s", id))
			return fmt.Errorf("Failed")
		default:
			panic(fmt.Sprintf("Illegal state: %s", job.Status))
		}
	}
	return nil
}

func createPackage(in string) (string, error) {
	out, err := ioutil.TempFile("", "*.zip")
	if err != nil {
		return "", err
	}
	return out.Name(), kuhnuri.Zip(out.Name(), in)
}

func (c *Client) doCreate(create *models.Create) (*models.Job, error) {
	b, err := json.Marshal(create)
	if err != nil {
		return nil, err
	}
	resp, err := http.Post(c.api("job"), "application/zip", bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var res models.Job
	err = json.Unmarshal(body, &res)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

func (c *Client) getJob(id string) (*models.Job, error) {
	resp, err := http.Get(c.api("job", id))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var res models.Job
	err = json.Unmarshal(body, &res)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

func (c *Client) getUpload() (*models.Upload, error) {
	resp, err := http.Get(c.api("upload"))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var res models.Upload
	err = json.Unmarshal(body, &res)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

func (c *Client) api(path ...string) string {
	return fmt.Sprintf("%s/%s", c.base, strings.Join(path, "/"))
}

func doUpload(zip string, url string) error {
	stat, err := os.Stat(zip)
	if err != nil {
		return err
	}
	buf, err := os.Open(zip)
	if err != nil {
		return err
	}
	client := &http.Client{}
	req, err := http.NewRequest("PUT", url, buf)
	req.ContentLength = stat.Size()
	resp, err := client.Do(req)
	defer resp.Body.Close()
	if err != nil {
		return err
	}
	return err
}
