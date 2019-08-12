package main

import (
	"github.com/kuhnuri/kuhnuri-cli/client"
	"github.com/kuhnuri/kuhnuri-cli/config"
	"gopkg.in/urfave/cli.v1"
	"io/ioutil"
	"log"
	"os"
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
	app.Author = "Jarno Elovirta <jarno@elovirta.com>"
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
		client, err := client.New(conf, transtype, input, output)
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
