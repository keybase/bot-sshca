package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/keybase/bot-ssh-ca/keybaseca/sshutils"

	"github.com/keybase/bot-ssh-ca/keybaseca/bot"
	"github.com/keybase/bot-ssh-ca/keybaseca/config"
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
			Value: config.DefaultConfigLocation,
			Usage: "Load configuration from `FILE`",
		},
	}
	app.Commands = []cli.Command{
		{
			Name:  "generate",
			Usage: "Generate a new CA key",
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name: "overwrite-existing-key",
				},
			},
			Action: func(c *cli.Context) error {
				conf, err := loadServerConfigAndWriteClientConfig(c.GlobalString("config"))
				if err != nil {
					return err
				}
				err = sshutils.Generate(conf, c.Bool("overwrite-existing-key"), true)
				if err != nil {
					return fmt.Errorf("Failed to generate a new key: %v", err)
				}
				return nil
			},
		},
		{
			Name:  "service",
			Usage: "Start the CA service in the foreground",
			Action: func(c *cli.Context) error {
				conf, err := loadServerConfigAndWriteClientConfig(c.GlobalString("config"))
				if err != nil {
					return err
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

// Write a kssh config file to /keybase/team/teamname.ssh/kssh-client.config. kssh will automatically pick up
// and use this config
func writeClientConfig(conf config.Config) error {
	// We only write the client config into the first team since that is enough for kssh to find it. This means
	// kssh will talk to the bot in the first team that is listed in the config file
	filename := filepath.Join("/keybase/team/", conf.GetTeams()[0], shared.ConfigFilename)
	username, err := bot.GetUsername(conf)
	if err != nil {
		return err
	}

	content, err := json.Marshal(kssh.ConfigFile{TeamName: conf.GetTeams()[0], BotName: username})

	return KBFSWrite(filename, string(content))
}

func KBFSWrite(filename string, contents string) error {
	cmd := exec.Command("keybase", "fs", "write", filename)
	cmd.Stdin = strings.NewReader(string(contents))
	bytes, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("Failed to write to file at %s: %s (%v)", filename, string(bytes), err)
	}
	return nil
}

func loadServerConfigAndWriteClientConfig(configFilename string) (config.Config, error) {
	if _, err := os.Stat(configFilename); os.IsNotExist(err) {
		return nil, fmt.Errorf("Config file at %s does not exist", configFilename)
	}
	conf, err := config.LoadConfig(configFilename)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse config file: %v", err)
	}
	err = writeClientConfig(conf)
	if err != nil {
		return nil, fmt.Errorf("Failed to write the client config: %v", err)
	}
	return conf, nil
}
