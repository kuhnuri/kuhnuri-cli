package main

import (
	"encoding/json"
	"github.com/kuhnuri/kuhnuri-cli/client"
	"github.com/kuhnuri/kuhnuri-cli/config"
	"github.com/kuhnuri/kuhnuri-cli/models"
	"github.com/urfave/cli"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"path/filepath"
)

var conf config.Config

func init() {
	conf = config.Read()
}

func main() {
	var transtype string
	var input string
	var output string
	var api string
	var project string
	var deliverable string

	app := cli.NewApp()
	app.Name = "kuhnuri"
	app.Version = "0.1.0"
	app.Usage = "Run DITA-OT on the cloud"
	app.Authors = []cli.Author{cli.Author{
		Name:  "Jarno Elovirta",
		Email: "jarno@elovirta.com",
	}}
	app.UsageText = "kuhnuri [options...]"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name: "f, format",
			//Required:    true,
			Usage:       "DITA-OT transtype",
			Destination: &transtype,
		},
		cli.StringFlag{
			Name: "i, input",
			//Required:    true,
			Usage:       "DITA input file, directory, or URL",
			Destination: &input,
		},
		cli.StringFlag{
			Name:        "o, output",
			Usage:       "Output directory or URL",
			Destination: &output,
		},
		cli.StringFlag{
			Name:        "p, project",
			Usage:       "Project file",
			Destination: &project,
		},
		cli.StringFlag{
			Name:        "d, deliverable",
			Usage:       "Project deliverable",
			Destination: &deliverable,
		},
		cli.StringFlag{
			Name:        "api",
			Usage:       "Kuhnuri API URL",
			Value:       conf["api"],
			Destination: &api,
		},
	}
	app.Action = func(c *cli.Context) error {
		if len(api) != 0 {
			conf["api"] = api
		}
		var deliv []*models.Deliverable
		if len(project) != 0 {
			p, err := readProject(project)
			if err != nil {
				return cli.NewExitError("ERROR: "+err.Error(), 1)
			}
			if len(deliverable) != 0 {
				for _, d := range p.Deliverables {
					if d.Id == deliverable {
						deliv = []*models.Deliverable{ &d }
					}
					if deliv == nil {
						return cli.NewExitError("ERROR: deliverable "+deliverable+" not found", 1)
					}
				}
			} else {
				deliv = &p.Deliverables
			}
		} else {
			if len(input) == 0 || len(transtype) == 0 {
				cli.ShowAppHelpAndExit(c, 1)
			}
			i, err := toUrl(input)
			if err != nil {
				return cli.NewExitError("ERROR: "+err.Error(), 1)
			}
			o, err := toUrl(output)
			if err != nil {
				return cli.NewExitError("ERROR: "+err.Error(), 1)
			}
			deliv = models.NewDeliverable(i.String(), transtype, o.String())
		}
		if deliv != nil {
			client, err := client.New(conf, deliv)
			if err != nil {
				return cli.NewExitError("ERROR: "+err.Error(), 1)
			}
			err = client.Execute()
			if err != nil {
				return cli.NewExitError("ERROR: "+err.Error(), 1)
			}
		}
		return nil
	}
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func readProject(path string) (*models.Project, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	in, err := ioutil.ReadFile(abs)
	if err != nil {
		return nil, err
	}
	project := models.Project{}
	err = json.Unmarshal(in, &project)
	if err != nil {
		return nil, err
	}
	return &project, nil
}

func toUrl(in string) (*url.URL, error) {
	uri, err := url.Parse(in)
	if err != nil {
		abs, err := filepath.Abs(in)
		if err != nil {
			return nil, err
		}
		return &url.URL{
			"file",
			"",
			nil,
			"",
			"",
			filepath.ToSlash(abs),
			false,
			"",
			"",
		}, nil
	} else if uri.IsAbs() {
		return uri, nil
	} else {
		abs, err := filepath.Abs(uri.Path)
		if err != nil {
			return nil, err
		}
		return &url.URL{
			"file",
			uri.Opaque,
			uri.User,
			uri.Host,
			abs,
			"",
			uri.ForceQuery,
			uri.RawQuery,
			uri.Fragment,
		}, nil
	}
}
