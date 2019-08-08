package client

import (
	"github.com/kuhnuri/kuhnuri-cli/client"
	"gopkg.in/urfave/cli.v1"
	"log"
	"os"
)

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
		client := client.New(transtype, input, output)
		return client.Execute()
	}
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
