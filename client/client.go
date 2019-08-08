package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	kuhnuri "github.com/kuhnuri/go-worker"
	"github.com/kuhnuri/kuhnuri-cli/models"
	"github.com/kuhnuri/kuhnuri-cli/spinner"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Client struct {
	transtype string
	input     string
	output    string
	spinner   *spinner.Spinner
}

func New(transtype string, input string, output string) *Client {
	return &Client{transtype, input, output, nil}
}

func (c *Client) Execute() error {
	c.spinner = spinner.New("Running")
	defer c.spinner.Stop()
	var in string
	if isLocal(c.input) {
		c.spinner.Message(fmt.Sprintf("Packaging %s", c.input))
		zip, err := createPackage(c.input)
		if err != nil {
			return fmt.Errorf("Failed to package source: %v", err)
		}
		c.spinner.Message("Uploading")
		upload, err := getUpload()
		if err != nil {
			return fmt.Errorf("Failed to retrieve upload URL: %v", err)
		}
		err = doUpload(zip, upload.Upload)
		if err != nil {
			return fmt.Errorf("Failed to upload source: %v", err)
		}
		in = upload.Url
	} else {
		in = c.input
	}
	c.spinner.Message("Submitting")
	create := models.NewCreate([]string{c.transtype}, in, c.output)
	job, err := doCreate(create)
	if err != nil {
		return fmt.Errorf("Failed to submit conversion: %v", err)
	}
	err = c.await(job)
	if err != nil {
		return err
	}
	if !isExternal(c.output) {
		c.spinner.Message(fmt.Sprintf("Downloading %s", job.Output))
		dst, err := doDownload(job.Output)
		if err != nil {
			return fmt.Errorf("Failed to download results: %v", err)
		}
		c.spinner.Message(fmt.Sprintf("Unpackaging %s", dst))
		kuhnuri.Unzip(dst, c.output)
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

func doDownload(in string) (string, error) {
	// if S3, get download URL
	//out, err := ioutil.TempFile("", "*.zip")
	//if err != nil {
	//	return "", err
	//}
	//s := api("job", id)
	//resp, err := http.Get(s)
	//if err != nil {
	//	return nil, err
	//}
	//defer resp.Body.Close()
	//body, err := ioutil.ReadAll(resp.Body)
	//if err != nil {
	//	return nil, err
	//}
	return "", nil
}

func (c *Client) await(created *models.Job) error {
	id := created.Id

	c.spinner.Message(fmt.Sprintf("Converting %s", id))
	ticks := time.Tick(5 * time.Second)
	for range ticks {
		job, err := getJob(id)
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
			c.spinner.Message(fmt.Sprintf("Done %s", id))
			return nil
		case "error":
			c.spinner.Message(fmt.Sprintf("Failed %s", id))
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

func doCreate(create *models.Create) (*models.Job, error) {
	b, err := json.Marshal(create)
	if err != nil {
		return nil, err
	}
	resp, err := http.Post(api("job"), "application/zip", bytes.NewReader(b))
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

func getJob(id string) (*models.Job, error) {
	s := api("job", id)
	resp, err := http.Get(s)
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

func getUpload() (*models.Upload, error) {
	resp, err := http.Get(api("upload"))
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

func api(path ...string) string {
	return fmt.Sprintf("https://3wfsdjcsf2.execute-api.eu-west-1.amazonaws.com/prod/api/v1/%s", strings.Join(path, "/"))
}

func doUpload(zip string, url string) error {
	buf, err := os.Open(zip)
	if err != nil {
		return err
	}
	resp, err := http.Post(url, "application/octet-stream", buf)
	defer resp.Body.Close()
	return err
}
