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
	input     string
	output    string
	spinner   *spinner.Spinner
}

func New(conf config.Config, transtype string, input string, output string) (*Client, error) {
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
	var in string
	if isLocal(c.input) {
		stat, err := os.Stat(c.input)
		if err != nil {
			return err
		}
		var dir string
		var start string
		if stat.IsDir() {
			dir = c.input
			filepath.Walk(c.input, func(path string, info os.FileInfo, err error) error {
				if filepath.Ext(path) == ".ditamap" && start == "" {
					start, err = filepath.Rel(c.input, path)
					if err != nil {
						return fmt.Errorf("Failed to build relative path for %s", path)
					}
					return filepath.SkipDir
				}
				return nil
			})
		} else {
			dir = filepath.Dir(c.input)
			start, err = filepath.Rel(dir, c.input)
			if err != nil {
				return fmt.Errorf("Failed to build relative path for %s", c.input)
			}
		}
		c.spinner.Message(fmt.Sprintf("Packaging"))
		zip, err := createPackage(dir)
		if err != nil {
			return fmt.Errorf("Failed to package source: %v", err)
		}
		c.spinner.Message("Uploading")
		upload, err := c.getUpload()
		if err != nil {
			return fmt.Errorf("Failed to retrieve upload URL: %v", err)
		}
		err = doUpload(zip, upload.Upload)
		if err != nil {
			return fmt.Errorf("Failed to upload source: %v", err)
		}
		in = fmt.Sprintf("jar:%s!/%s", upload.Url, filepath.ToSlash(start))
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
		c.spinner.Message(fmt.Sprintf("Downloading"))
		dst, err := c.doDownload(job.Id)
		if err != nil {
			return fmt.Errorf("Failed to download results: %v", err)
		}
		c.spinner.Message(fmt.Sprintf("Unpackaging %s", dst))
		err = kuhnuri.MkDirs(c.output)
		if err != nil {
			return err
		}
		err = kuhnuri.Unzip(dst, c.output)
		if err != nil {
			return fmt.Errorf("Failed to unpackage %s: %v", dst, err)
		}
	}

	return nil
}

func isExternal(in string) bool {
	url, err := url.Parse(in)
	if err != nil {
		return false
	}
	if !url.IsAbs() {
		return false
	}
	if url.Scheme == "file" {
		return false
	}
	return true
}

func isLocal(in string) bool {
	abs, err := filepath.Abs(in)
	if err != nil {
		return false
	}
	_, err = os.Stat(abs)
	if os.IsNotExist(err) {
		return false
	}
	return true
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
