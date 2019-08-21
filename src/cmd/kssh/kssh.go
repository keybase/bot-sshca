package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/keybase/bot-ssh-ca/src/keybaseca/sshutils"

	"github.com/google/uuid"
	"github.com/keybase/bot-ssh-ca/src/kssh"
	"github.com/keybase/bot-ssh-ca/src/shared"

	"golang.org/x/crypto/ssh"
)

func main() {
	team, remainingArgs, action, runtimeConfig, err := handleArgs(os.Args[1:])
	if err != nil {
		fmt.Printf("Failed to parse arguments: %v\n", err)
		os.Exit(1)
	}
	keyPath, err := getSignedKeyLocation(team)
	if err != nil {
		fmt.Printf("Failed to retrieve location to store SSH keys: %v\n", err)
		os.Exit(1)
	}
	if isValidCert(keyPath) {
		kssh.DebugLog(runtimeConfig, "Reusing unexpired certificate")
		doAction(runtimeConfig, action, keyPath, remainingArgs)
	}
	config, err := getConfig(team)
	if err != nil {
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}
	err = provisionNewKey(runtimeConfig, config, keyPath)
	if err != nil {
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}
	doAction(runtimeConfig, action, keyPath, remainingArgs)
}

func doAction(runtimeConfig kssh.RuntimeConfig, action Action, keyPath string, remainingArgs []string) {
	if action == SSH {
		runSSHWithKey(runtimeConfig, keyPath, remainingArgs)
	} else if action == Provision {
		err := kssh.AddKeyToSSHAgent(keyPath)
		if err != nil {
			fmt.Printf("%v\n", err)
			os.Exit(1)
		}
		kssh.DebugLog(runtimeConfig, "Provisioned new SSH key at %s\n", keyPath)
	}
}

// getSignedKeyLocation returns the path of where the signed SSH key should be stored. botname is the name of the bot
// specified via --bot if specified. It is necessary to include the bot in the filename in order to properly
// handle how the switch bot flow interacts with the isValidCert function
func getSignedKeyLocation(botname string) (string, error) {
	signedKeyLocation := shared.ExpandPathWithTilde("~/.ssh/keybase-signed-key--")
	if botname != "" {
		return signedKeyLocation + botname, nil
	}
	defaultBot, _, err := kssh.GetDefaultBotAndTeam()
	if err != nil {
		return "", err
	}
	return signedKeyLocation + defaultBot, nil
}

var cliArguments = []kssh.CLIArgument{
	{Name: "--set-default-bot", HasArgument: true},
	{Name: "--clear-default-bot", HasArgument: false},
	{Name: "--bot", HasArgument: true},
	{Name: "--provision", HasArgument: false},
	{Name: "--set-default-user", HasArgument: true},
	{Name: "--clear-default-user", HasArgument: false},
	{Name: "--help", HasArgument: false},
	{Name: "-v", HasArgument: false, Preserve: true},
}

func generateHelpPage() string {
	return fmt.Sprintf(`NAME:
   kssh - A replacement ssh binary using Keybase SSH CA to provision SSH keys

USAGE:
   kssh [kssh options] [ssh arguments...]

VERSION:
   0.0.1

GLOBAL OPTIONS:
   --help                Show help
   -v                    Enable kssh and ssh debug logs
   --provision           Provision a new SSH key and add it to the ssh-agent. Useful if you need to run another 
                         program that uses SSH auth (eg scp, rsync, etc)
   --set-default-bot     Set the default bot to be used for kssh. Not necessary if you are only in one team that
                         is using Keybase SSH CA
   --clear-default-bot   Clear the default bot
   --bot                 Specify a specific bot to be used for kssh. Not necessary if you are only in one team that
                         is using Keybase SSH CA
   --set-default-user    Set the default SSH user to be used for kssh. Useful if you use ssh configs that do not set 
					     a default SSH user 
   --clear-default-user  Clear the default SSH user`)
}

type Action int

const (
	Provision Action = iota
	SSH
)

// Returns botname, remaining arguments, action, runtimeConfig, error
// If the argument requires exiting after processing, it will call os.Exit
func handleArgs(args []string) (string, []string, Action, kssh.RuntimeConfig, error) {
	remaining, found, err := kssh.ParseArgs(args, cliArguments)
	if err != nil {
		return "", nil, 0, kssh.RuntimeConfig{}, fmt.Errorf("Failed to parse provided arguments: %v", err)
	}

	debug := false
	team := ""
	action := SSH
	for _, arg := range found {
		if arg.Argument.Name == "--bot" {
			team = arg.Value
		}
		if arg.Argument.Name == "--set-default-user" {
			err := kssh.SetDefaultSSHUser(arg.Value)
			if err != nil {
				fmt.Printf("Failed to set the default ssh user: %v\n", err)
				os.Exit(1)
			}
			fmt.Println("Set default ssh user, exiting...")
			os.Exit(0)
		}
		if arg.Argument.Name == "--clear-default-user" {
			err := kssh.SetDefaultSSHUser("")
			if err != nil {
				fmt.Printf("Failed to clear the default ssh user: %v\n", err)
				os.Exit(1)
			}
			fmt.Println("Cleared default ssh user, exiting...")
			os.Exit(0)
		}
		if arg.Argument.Name == "--set-default-bot" {
			// We exit immediately after setting the default bot
			err := kssh.SetDefaultBot(arg.Value)
			if err != nil {
				fmt.Printf("Failed to set the default bot: %v\n", err)
				os.Exit(1)
			}
			fmt.Println("Set default bot, exiting...")
			os.Exit(0)
		}
		if arg.Argument.Name == "--clear-default-bot" {
			err := kssh.SetDefaultBot("")
			if err != nil {
				fmt.Printf("Failed to clear the default bot: %v\n", err)
				os.Exit(1)
			}
			fmt.Println("Cleared default bot, exiting...")
			os.Exit(0)
		}
		if arg.Argument.Name == "--provision" {
			action = Provision
		}
		if arg.Argument.Name == "--help" {
			fmt.Println(generateHelpPage())
			os.Exit(0)
		}
		if arg.Argument.Name == "-v" {
			debug = true
		}
	}
	return team, remaining, action, kssh.RuntimeConfig{Debug: debug}, nil
}

// Get the kssh.ConfigFile. botname is the team specified via --bot if one was specified, otherwise the empty string
func getConfig(botname string) (kssh.ConfigFile, error) {
	empty := kssh.ConfigFile{}

	// They specified a bot via `kssh --bot cabot ...`
	if botname != "" {
		team, err := kssh.GetTeamFromBot(botname)
		if err != nil {
			return empty, err
		}
		conf, err := kssh.LoadConfig(fmt.Sprintf("/keybase/team/%s/%s", team, shared.ConfigFilename))
		if err != nil {
			return empty, fmt.Errorf("Failed to load config file for team=%s: %v", team, err)
		}
		return conf, nil
	}

	// Check if they set a default bot and retrieve the config for that bot/team if so
	defaultBot, defaultTeam, err := kssh.GetDefaultBotAndTeam()
	if err != nil {
		return empty, err
	}
	if defaultBot != "" && defaultTeam != "" {
		conf, err := kssh.LoadConfig(fmt.Sprintf("/keybase/team/%s/%s", defaultTeam, shared.ConfigFilename))
		if err != nil {
			return empty, fmt.Errorf("Failed to load config file for default bot=%s, team=%s: %v", defaultBot, defaultTeam, err)
		}
		return conf, nil
	}

	// No specified bot and no default bot, fallback and load all the configs
	configs, botnames, err := kssh.LoadConfigs()
	if err != nil {
		return empty, fmt.Errorf("Failed to load config file(s): %v\n", err)
	}
	if len(configs) == 0 {
		return empty, fmt.Errorf("Did not find any config files in KBFS (is `keybaseca service` running?)")
	} else if len(configs) == 1 {
		return configs[0], nil
	} else {
		noDefaultTeamMessage := fmt.Sprintf("Found %v config files (%s). No default bot is configured. \n"+
			"Either specify a team via `kssh --bot cabotname` or set a default bot via `kssh --set-default-bot cabotname`", len(configs), strings.Join(botnames, ", "))
		return empty, fmt.Errorf(noDefaultTeamMessage)
	}
}

// Returns whether or not the cert at the given path is a valid unexpired certificate
func isValidCert(keyPath string) bool {
	_, err1 := os.Stat(keyPath)
	_, err2 := os.Stat(shared.KeyPathToPubKey(keyPath))
	_, err3 := os.Stat(shared.KeyPathToCert(keyPath))
	if os.IsNotExist(err1) || os.IsNotExist(err2) || os.IsNotExist(err3) {
		return false // Cert does not exist
	}

	certBytes, err := ioutil.ReadFile(shared.KeyPathToCert(keyPath))
	if err != nil {
		// Failed to read the file for some reason, just provision a new cert
		return false
	}
	k, _, _, _, err := ssh.ParseAuthorizedKey(certBytes)
	if err != nil {
		// Failed to parse it so just provision a new cert
		return false
	}
	// This is legal, see: https://github.com/golang/go/issues/22046
	cert := k.(*ssh.Certificate)
	validBefore := time.Unix(int64(cert.ValidBefore), 0)
	validAfter := time.Unix(int64(cert.ValidAfter), 0)
	return time.Now().After(validAfter) && time.Now().Before(validBefore)
}

// Provision a new signed SSH key with the given config
func provisionNewKey(runtimeConfig kssh.RuntimeConfig, config kssh.ConfigFile, keyPath string) error {
	kssh.DebugLog(runtimeConfig, "Generating a new SSH key...")
	err := sshutils.GenerateNewSSHKey(keyPath, true, false)
	if err != nil {
		return fmt.Errorf("Failed to generate a new SSH key: %v", err)
	}
	pubKey, err := ioutil.ReadFile(shared.KeyPathToPubKey(keyPath))
	if err != nil {
		return fmt.Errorf("Failed to read the SSH key from the filesystem: %v", err)
	}

	randomUUID, err := uuid.NewRandom()
	if err != nil {
		return fmt.Errorf("Failed to generate a new UUID for the SignatureRequest: %v", err)
	}

	kssh.DebugLog(runtimeConfig, "Requesting signature from the CA....")
	resp, err := kssh.GetSignedKey(config, shared.SignatureRequest{
		UUID:         randomUUID.String(),
		SSHPublicKey: string(pubKey),
	})
	if err != nil {
		return fmt.Errorf("Failed to get a signed key from the CA: %v", err)
	}
	kssh.DebugLog(runtimeConfig, "Received signature from the CA!")

	err = ioutil.WriteFile(shared.KeyPathToCert(keyPath), []byte(resp.SignedKey), 0600)
	if err != nil {
		return fmt.Errorf("Failed to write new SSH key to disk: %v", err)
	}

	return nil
}

// Run SSH with the given key. Calls os.Exit and does not return.
func runSSHWithKey(runtimeConfig kssh.RuntimeConfig, keyPath string, remainingArgs []string) {
	// Determine whether a default SSH user has been specified and configure it if so
	useConfig := false
	user, err := kssh.GetDefaultSSHUser()
	if err != nil {
		fmt.Printf("Failed to retrieve default SSH user: %v\n", err)
		os.Exit(1)
	}
	if user != "" {
		useConfig = true
		err = kssh.CreateDefaultUserConfigFile(keyPath)
		if err != nil {
			fmt.Printf("Failed to set default user: %v\n", err)
			os.Exit(1)
		}
	}

	// Add the key to the ssh-agent in case we are doing multiple connections (eg via the `-J` flag)
	err = kssh.AddKeyToSSHAgent(keyPath)
	if err != nil {
		fmt.Printf("Failed to add SSH key to the SSH agent: %v\n", err)
		os.Exit(1)
	}

	argumentList := []string{"-i", keyPath, "-o", "IdentitiesOnly=yes"}
	checkAndWarnOnUnspecifiedBehavior(useConfig, remainingArgs)
	if useConfig {
		argumentList = append(argumentList, "-F", kssh.AlternateSSHConfigFile)
		kssh.DebugLog(runtimeConfig, "Using default ssh user %s", user)
	}

	argumentList = append(argumentList, remainingArgs...)

	cmd := exec.Command("ssh", argumentList...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	err = cmd.Run()

	if err != nil {
		fmt.Printf("SSH exited with err: %v\n", err)
		os.Exit(1)
	}
	os.Exit(0)
}

func checkAndWarnOnUnspecifiedBehavior(useConfig bool, arguments []string) {
	if useConfig {
		for _, arg := range arguments {
			if arg == "-F" {
				fmt.Println("Warning: You passed a -F flag, but kssh also uses this argument in " +
					"order to implement support for a default SSH username, which you're also using. " +
					"Either do not use the -F flag or run `kssh --clear-default-user` to reset the " +
					"default SSH user and delegate this to the running CA bot.")
			}
		}
	}
}
