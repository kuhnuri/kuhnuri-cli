package main

import (
	"github.com/kuhnuri/kuhnuri-cli/client"
	"github.com/kuhnuri/kuhnuri-cli/config"
	"github.com/urfave/cli"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"path/filepath"
)

var conf config.Config

func init() {
	log.SetOutput(ioutil.Discard)
	conf = config.Read()
}

func main() {
	var transtype string
	var input string
	var output string
	var api string

	app := cli.NewApp()
	app.Name = "kuhnuri"
	app.Version = "0.1.0"
	app.Usage = "Run DITA-OT on the cloud"
	app.Authors = []cli.Author{cli.Author{
		Name: "Jarno Elovirta",
		Email: "jarno@elovirta.com",
	}}
	app.UsageText = "kuhnuri [options...]"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "f, format",
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
			Usage:       "Output directory or URL",
			Destination: &output,
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
		i, err := toUrl(input)
		if err != nil {
			return cli.NewExitError("ERROR: "+err.Error(), 1)
		}
		o, err := toUrl(output)
		if err != nil {
			return cli.NewExitError("ERROR: "+err.Error(), 1)
		}
		client, err := client.New(conf, transtype, i, o)
		if err != nil {
			return cli.NewExitError("ERROR: "+err.Error(), 1)
		}
		err = client.Execute()
		if err != nil {
			return cli.NewExitError("ERROR: "+err.Error(), 1)
		}
		return nil
	}
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
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
