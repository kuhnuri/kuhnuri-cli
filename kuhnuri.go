package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	kuhnuri "github.com/kuhnuri/go-worker"
	"gopkg.in/urfave/cli.v1"
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
	spinner   *Spinner
}

func (c *Client) Execute() error {
	c.spinner = NewSpinner(fmt.Sprintf("Collecting"))
	zip, err := createPackage()
	if err != nil {
		return err
	}
	c.spinner.msg = fmt.Sprintf("Uploading")
	upload, err := getUpload()
	if err != nil {
		return err
	}
	doUpload(zip, upload.Upload)

	c.spinner.msg = fmt.Sprintf("Submitting")
	create := NewCreate([]string{c.transtype}, upload.Url, c.output)
	job, err := doCreate(create)
	if err != nil {
		return err
	}
	id := job.Id
	status := job.Status

	c.spinner.msg = fmt.Sprintf("Converting %s %s", id, status)
	ticks := time.Tick(5 * time.Second)
	for range ticks {
		job, err = getJob(id)
		if err != nil {
			return err
		}
		switch job.Status {
		case "queue":
			fallthrough
		case "process":
			c.spinner.msg = fmt.Sprintf("Converting %s %s", id, job.Status)
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

func doCreate(create *Create) (*Job, error) {
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
	var res Job
	err = json.Unmarshal(body, &res)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

func getJob(id string) (*Job, error) {
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
	var res Job
	err = json.Unmarshal(body, &res)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

func getUpload() (*Upload, error) {
	resp, err := http.Get(api("upload"))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var res Upload
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

func main() {
	var transtype string
	var input string
	var output string

	app := cli.NewApp()
	app.Name = "kuhnuri"
	app.Version = "0.1.0"
	app.Usage = "Run DITA-OT on the cloud"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "t, transtype",
			Required:    true,
			Usage:       "DITA-OT transtype",
			Destination: &transtype,
		},
		cli.StringFlag{
			Name:        "i, input",
			Required:    true,
			Usage:       "DITA input file, directory, or URL",
			Destination: &input,
		},
		cli.StringFlag{
			Name:        "o, output",
			Usage:       "Output file, directory, or URL",
			Destination: &output,
		},
	}
	app.Action = func(c *cli.Context) error {
		client := Client{transtype, input, output, nil}
		err := client.Execute()
		fmt.Printf("Done: %s", err)
		return err
	}
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
