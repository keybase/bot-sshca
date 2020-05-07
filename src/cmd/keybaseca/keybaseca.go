package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/keybase/bot-sshca/src/keybaseca/ca"
	"github.com/keybase/bot-sshca/src/keybaseca/constants"

	"github.com/google/uuid"

	"github.com/keybase/bot-sshca/src/keybaseca/config"
	klog "github.com/keybase/bot-sshca/src/keybaseca/log"
	"github.com/keybase/bot-sshca/src/keybaseca/sshutils"
	"github.com/keybase/bot-sshca/src/shared"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

var VersionNumber = "master"

func main() {
	app := cli.NewApp()
	app.Name = "keybaseca"
	app.Usage = "An SSH CA built on top of Keybase"
	app.Version = VersionNumber
	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:  "debug",
			Usage: "Log debug information",
		},
		cli.BoolFlag{
			Name:   "wipe-all-configs",
			Hidden: true,
			Usage:  "Clean all client configs the CA Keybase user can find from KV stores",
		},
		cli.BoolFlag{
			Name:   "wipe-logs",
			Hidden: true,
			Usage:  "Used in the integration tests to delete all CA logs",
		},
	}
	app.Commands = []cli.Command{
		{
			Name:   "backup",
			Usage:  "Print the current CA private key to stdout for backup purposes",
			Action: backupAction,
			Before: beforeAction,
		},
		{
			Name:   "generate",
			Usage:  "Generate a new CA key",
			Action: generateAction,
			Before: beforeAction,
		},
		{
			Name:   "service",
			Usage:  "Start the CA service in the foreground",
			Action: serviceAction,
			Before: beforeAction,
		},
		{
			Name:  "sign",
			Usage: "Sign a given public key with all permissions without a dependency on Keybase",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:     "public-key",
					Usage:    "The path to the public key you wish to sign. Eg `~/.ssh/id_rsa.pub`",
					Required: true,
				},
				cli.BoolFlag{
					Name:  "overwrite",
					Usage: "Overwrite the existing certificate on the filesystem",
				},
			},
			Action: signAction,
			Before: beforeAction,
		},
	}
	app.Action = mainAction
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

// The action for the `keybaseca backup` subcommand
func backupAction(c *cli.Context) error {
	fmt.Println("Are you sure you want to export the CA private key? If this key is compromised, an " +
		"attacker could access every server that you have configured with this bot. Type \"yes\" to confirm.")
	var response string
	_, err := fmt.Scanln(&response)
	if err != nil {
		return err
	}
	if response != "yes" {
		return fmt.Errorf("Did not get confirmation of key export, aborting")
	}

	conf, err := loadServerConfig()
	if err != nil {
		return err
	}
	bytes, err := ioutil.ReadFile(conf.GetCAKeyLocation())
	if err != nil {
		return fmt.Errorf("Failed to load the CA key from %s: %v", conf.GetCAKeyLocation(), err)
	}
	klog.Log(conf, "Exported CA key to stdout")
	fmt.Println("\nKeep this key somewhere very safe. We recommend keeping a physical copy of it in a secure place.")
	fmt.Println("")
	fmt.Println(string(bytes))
	return nil
}

// The action for the `keybaseca generate` subcommand
func generateAction(c *cli.Context) error {
	conf, err := loadServerConfig()
	if err != nil {
		return err
	}
	err = sshutils.Generate(conf, strings.ToLower(os.Getenv("FORCE_WRITE")) == "true")
	if err != nil {
		return fmt.Errorf("Failed to generate a new key: %v", err)
	}
	return nil
}

// The action for the `keybaseca service` subcommand
func serviceAction(c *cli.Context) error {
	conf, err := loadServerConfig()
	if err != nil {
		return err
	}
	err = startCA(conf)
	if err != nil {
		return fmt.Errorf("CA chatbot crashed: %v", err)
	}
	return nil
}

func startCA(conf config.Config) error {
	cabot, err := ca.New(conf)
	if err != nil {
		return err
	}
	fmt.Println("Starting CA bot...")
	return cabot.Start()
}

// The action for the `keybaseca sign` subcommand
func signAction(c *cli.Context) error {
	// Skip validation of the config since that relies on Keybase's servers
	conf := config.EnvConfig{}
	err := config.ValidateConfig(conf, true)
	if err != nil {
		return fmt.Errorf("Invalid config: %v", err)
	}
	principals := strings.Join(conf.GetTeams(), ",")
	expiration := conf.GetKeyExpiration()
	randomUUID, err := uuid.NewRandom()
	if err != nil {
		return fmt.Errorf("Failed to generate unique key ID: %v", err)
	}

	// Read the public key from the specified file
	filename := c.String("public-key")
	pubKey, err := ioutil.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("Failed to read file at %s to get the public key: %v", filename, err)
	}

	// Sign the public key
	signature, err := sshutils.SignKey(conf.GetCAKeyLocation(), randomUUID.String()+":keybaseca-sign", principals, expiration, string(pubKey))
	if err != nil {
		return fmt.Errorf("Failed to sign key: %v", err)
	}

	// Either store it in a file or print it to stdout
	certPath := shared.KeyPathToCert(shared.PubKeyPathToKeyPath(filename))
	_, err = os.Stat(certPath)
	if os.IsNotExist(err) || c.Bool("overwrite") {
		err = ioutil.WriteFile(certPath, []byte(signature), 0600)
		if err != nil {
			return fmt.Errorf("Failed to write certificate to file: %v", err)
		}
		fmt.Printf("Provisioned new certificate in %s\n", certPath)
	} else {
		fmt.Printf("Provisioned new certificate. Place this in %s in order to use it with ssh.\n", certPath)
		fmt.Printf("\n```\n%s```\n", signature)
	}
	return nil
}

// A global before action that handles the --debug flag by setting the logrus logging level
func beforeAction(c *cli.Context) error {
	if c.GlobalBool("debug") {
		logrus.SetLevel(logrus.DebugLevel)
	}
	return nil
}

// The action for the `keybaseca` command. Only used for hidden and unlisted flags.
func mainAction(c *cli.Context) error {
	switch {
	case c.Bool("wipe-all-configs"):
		conf, err := loadServerConfig()
		if err != nil {
			return err
		}
		if err = deleteAllClientConfigs(conf); err != nil {
			return err
		}
	case c.Bool("wipe-logs"):
		conf, err := loadServerConfig()
		if err != nil {
			return err
		}
		logLocation := conf.GetLogLocation()
		if strings.HasPrefix(logLocation, "/keybase/") {
			err = constants.GetDefaultKBFSOperationsStruct().Delete(logLocation)
			if err != nil {
				return fmt.Errorf("Failed to delete log file at %s: %v", logLocation, err)
			}
		} else {
			err = os.Remove(logLocation)
			if err != nil {
				return fmt.Errorf("Failed to delete log file at %s: %v", logLocation, err)
			}
		}
		fmt.Println("Wiped existing log file at " + logLocation)
	default:
		cli.ShowAppHelpAndExit(c, 1)
	}
	return nil
}

func deleteAllClientConfigs(conf config.Config) error {
	cabot, err := ca.New(conf)
	if err != nil {
		return err
	}
	return cabot.DeleteAllClientConfigs()
}

// Load and validate a server config object from the environment
func loadServerConfig() (config.Config, error) {
	conf := config.EnvConfig{}
	err := config.ValidateConfig(conf, false)
	if err != nil {
		return nil, fmt.Errorf("Failed to validate config: %v", err)
	}
	return &conf, nil
}
