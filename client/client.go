package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	kuhnuri "github.com/kuhnuri/go-worker"
	"github.com/kuhnuri/kuhnuri-cli/models"
	"github.com/kuhnuri/kuhnuri-cli/spinner"
	"io/ioutil"
	"log"
	"net/http"
	"os"
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
	c.spinner = spinner.New("Collecting")
	zip, err := createPackage()
	if err != nil {
		return err
	}
	c.spinner.Msg = "Uploading"
	upload, err := getUpload()
	if err != nil {
		return err
	}
	doUpload(zip, upload.Upload)

	c.spinner.Msg = "Submitting"
	create := models.NewCreate([]string{c.transtype}, upload.Url, c.output)
	job, err := doCreate(create)
	if err != nil {
		return err
	}
	err = c.await(job)

	return err
}

func (c *Client) await(created *models.Job) error {
	id := created.Id
	status := created.Status

	c.spinner.Msg = fmt.Sprintf("Converting %s %s", id, status)
	ticks := time.Tick(5 * time.Second)
	for range ticks {
		job, err := getJob(id)
		if err != nil {
			return err
		}
		switch job.Status {
		case "queue":
			fallthrough
		case "process":
			c.spinner.Msg = fmt.Sprintf("Converting %s %s", id, job.Status)
			break
		case "done":
			fallthrough
		case "error":
			c.spinner.Stop()
			return nil
		default:
			log.Fatalf("Illegal state: %s", job.Status)
		}
	}
	return nil
}

func createPackage() (string, error) {
	out, err := ioutil.TempFile("", "*.zip")
	if err != nil {
		return "", err
	}
	return out.Name(), kuhnuri.Zip(out.Name(), "inputdir")
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
	resp, err := http.Post(url, "image/jpeg", buf)
	defer resp.Body.Close()
	return err
}
