package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/keybase/bot-ssh-ca/keybaseca/bot"
	"github.com/keybase/bot-ssh-ca/keybaseca/config"
	"github.com/keybase/bot-ssh-ca/keybaseca/libca"
	"github.com/keybase/bot-ssh-ca/keybaseca/sshutils"
	"github.com/keybase/bot-ssh-ca/kssh"
	"github.com/keybase/bot-ssh-ca/shared"

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
				err = sshutils.Generate(conf, c.Bool("overwrite-existing-key"), true)
				if err != nil {
					return fmt.Errorf("Failed to generate a new key: %v", err)
				}
				err = writeClientConfig(conf)
				if err != nil {
					return fmt.Errorf("Failed to write the client config!")
				}
				return nil
			},
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name: "overwrite-existing-key",
				},
			},
		},
		{
			Name:  "service",
			Usage: "Start the CA service in the foreground",
			Action: func(c *cli.Context) error {
				configFilename := c.GlobalString("config")
				if _, err := os.Stat(configFilename); os.IsNotExist(err) {
					return fmt.Errorf("Config file at %s does not exist", configFilename)
				}
				conf, err := config.LoadConfig(configFilename)
				if err != nil {
					return fmt.Errorf("Failed to parse config file: %v", err)
				}
				err = writeClientConfig(conf)
				if err != nil {
					return fmt.Errorf("Failed to write the client config!")
				}
				err = bot.StartBot(conf)
				if err != nil {
					return fmt.Errorf("CA chatbot crashed: %v", err)
				}
				return nil
			},
			Flags: []cli.Flag{},
		},
	}
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func writeClientConfig(conf config.Config) error {
	filename := filepath.Join("/keybase/team/", conf.GetTeamName(), shared.ConfigFilename)
	username, err := bot.GetUsername(conf)
	if err != nil {
		return err
	}

	content, err := json.Marshal(kssh.ConfigFile{TeamName: conf.GetTeamName(), BotName: username})

	return ioutil.WriteFile(filename, content, 0600)
}
