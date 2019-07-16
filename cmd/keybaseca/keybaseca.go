package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

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
		cli.BoolFlag{
			Name:   "wipe-all-configs",
			Hidden: true,
			Usage:  "Used in the integration tests to clean all client configs from KBFS",
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
				captureControlCToDeleteClientConfig(conf)
				defer deleteClientConfig(conf)
				err = sshutils.Generate(conf, c.Bool("overwrite-existing-key") || os.Getenv("FORCE_WRITE") == "true", true)
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
				captureControlCToDeleteClientConfig(conf)
				defer deleteClientConfig(conf)
				err = bot.StartBot(conf)
				if err != nil {
					return fmt.Errorf("CA chatbot crashed: %v", err)
				}
				return nil
			},
			Flags: []cli.Flag{},
		},
	}
	app.Action = func(c *cli.Context) error {
		if c.Bool("wipe-all-configs") {
			teams, err := shared.KBFSList("/keybase/team/")
			if err != nil {
				return err
			}

			semaphore := make(chan interface{}, len(teams))
			for _, team := range teams {
				go func(team string) {
					filename := fmt.Sprintf("/keybase/team/%s/%s", team, shared.ConfigFilename)
					exists, _ := shared.KBFSFileExists(filename)
					if exists {
						err = shared.KBFSDelete(filename)
						if err != nil {
							fmt.Printf("%v\n", err)
						}
					}
					semaphore <- 0
				}(team)
			}
			for i := 0; i < len(teams); i++ {
				<-semaphore
			}
		}
		return nil
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

	return shared.KBFSWrite(filename, string(content))
}

// Delete the client config file. Run when the CA bot is terminating so that KBFS does not contain any stale
// client config files
func deleteClientConfig(conf config.Config) error {
	filename := filepath.Join("/keybase/team/", conf.GetTeams()[0], shared.ConfigFilename)
	return shared.KBFSDelete(filename)
}

func captureControlCToDeleteClientConfig(conf config.Config) {
	signalChan := make(chan os.Signal)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-signalChan
		fmt.Println("losing CA bot...")
		err := deleteClientConfig(conf)
		if err != nil {
			fmt.Printf("Failed to delete client config: %v", err)
			os.Exit(1)
		}
		os.Exit(0)
	}()
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
