package main

import (
	"fmt"
	"github.com/keybase/bot-ssh-ca/keybaseca/config"
	"github.com/keybase/bot-ssh-ca/keybaseca/generate"
	"github.com/keybase/bot-ssh-ca/keybaseca/libca"
	"log"
	"os"

	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()
	app.Name = "keybaseca"
	app.Usage = "An SSH CA built on top of Keybase"
	app.Version = "0.0.1"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "config, c",
			Value: libca.ExpandPathWithTilde("~/keybaseca.config"),
			Usage: "Load configuration from `FILE`",
		},
	}
	app.Commands = []cli.Command{
		{
			Name:  "generate",
			Usage: "Generate a new CA key",
			Action: func(c *cli.Context) error {
				configFilename := c.GlobalString("config")
				if _, err := os.Stat(configFilename); os.IsNotExist(err) {
					return fmt.Errorf("Config file at %s does not exist", configFilename)
				}
				conf, err := config.LoadConfig(configFilename)
				if err != nil {
					return fmt.Errorf("Failed to parse config file: %v", err)
				}
				err = generate.Generate(conf, c.Bool("overwrite-existing-key"))
				if err != nil {
					return fmt.Errorf("Failed to generate a new key: %v", err)
				}
				return nil
			},
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name: "overwrite-existing-key",
				},
			},
		},
	}
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}

}
