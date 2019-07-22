package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/keybase/bot-ssh-ca/keybaseca/sshutils"

	"github.com/google/uuid"
	"github.com/keybase/bot-ssh-ca/kssh"
	"github.com/keybase/bot-ssh-ca/shared"

	"golang.org/x/crypto/ssh"
)

func main() {
	team, remainingArgs, action, err := handleArgs(os.Args[1:])
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
		doAction(action, keyPath, remainingArgs)
	}
	config, err := getConfig(team)
	if err != nil {
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}
	err = provisionNewKey(config, keyPath)
	if err != nil {
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}
	doAction(action, keyPath, remainingArgs)
}

func doAction(action Action, keyPath string, remainingArgs []string) {
	if action == SSH {
		runSSHWithKey(keyPath, remainingArgs)
	} else if action == Provision {
		err := kssh.AddKeyToSSHAgent(keyPath)
		if err != nil {
			fmt.Printf("%v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Provisioned new SSH key at %s", keyPath)
	}
}

// getSignedKeyLocation returns the path of where the signed SSH key should be stored. team is the name of the team
// specified via --team if specified. It is necessary to include the team in the filename in order to properly
// handle how the switch team flow interacts with the isValidCert function
func getSignedKeyLocation(team string) (string, error) {
	signedKeyLocation := shared.ExpandPathWithTilde("~/.ssh/keybase-signed-key--")
	if team != "" {
		return signedKeyLocation + team, nil
	}
	defaultTeam, err := kssh.GetDefaultTeam()
	if err != nil {
		return "", err
	}
	return signedKeyLocation + defaultTeam, nil
}

var cliArguments = []kssh.CLIArgument{
	{Name: "--set-default-team", HasArgument: true},
	{Name: "--team", HasArgument: true},
	{Name: "--provision", HasArgument: false},
}

type Action int

const (
	Provision Action = iota
	SSH
)

// Returns team, remaining arguments, action, error
func handleArgs(args []string) (string, []string, Action, error) {
	remaining, found, err := kssh.ParseArgs(args, cliArguments)
	if err != nil {
		return "", nil, 0, fmt.Errorf("Failed to parse provided arguments: %v", err)
	}

	team := ""
	action := SSH
	for _, arg := range found {
		if arg.Name == "--team" {
			team = arg.Value
		}
		if arg.Name == "--set-default-team" {
			// We exit immediately after setting the default team
			err := kssh.SetDefaultTeam(args[2])
			if err != nil {
				fmt.Printf("Failed to set the default team: %v", err)
				os.Exit(1)
			}
			fmt.Println("Set default team, exiting...")
			os.Exit(0)
		}
		if arg.Name == "--provision" {
			action = Provision
		}
	}
	return team, remaining, action, nil
}

// Get the kssh.ConfigFile. team is the team specified via --team if one was specified, otherwise the empty string
func getConfig(team string) (kssh.ConfigFile, error) {
	empty := kssh.ConfigFile{}

	// They specified a team via `kssh --team teamname.ssh ...`
	if team != "" {
		conf, err := kssh.LoadConfig(fmt.Sprintf("/keybase/team/%s/%s", team, shared.ConfigFilename))
		if err != nil {
			return empty, fmt.Errorf("Failed to load config file for team=%s: %v", team, err)
		}
		return conf, nil
	}

	// They set a default team
	defaultTeam, err := kssh.GetDefaultTeam()
	if err != nil {
		return empty, err
	}
	if defaultTeam != "" {
		conf, err := kssh.LoadConfig(fmt.Sprintf("/keybase/team/%s/%s", defaultTeam, shared.ConfigFilename))
		if err != nil {
			return empty, fmt.Errorf("Failed to load config file for team=%s: %v", defaultTeam, err)
		}
		return conf, nil
	}

	// No specified team and no default team, fallback and load all the configs
	configs, teams, err := kssh.LoadConfigs()
	if err != nil {
		return empty, fmt.Errorf("Failed to load config file(s): %v\n", err)
	}
	if len(configs) == 0 {
		return empty, fmt.Errorf("Did not find any config files in KBFS (is `keybaseca service` running?)")
	} else if len(configs) == 1 {
		return configs[0], nil
	} else {
		noDefaultTeamMessage := fmt.Sprintf("Found %v config files (%s). No default team is configured. \n"+
			"Either specify a team via `kssh --team teamname.ssh` or set a default team via `kssh --set-default-team teamname.ssh`", len(configs), strings.Join(teams, ", "))
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
func provisionNewKey(config kssh.ConfigFile, keyPath string) error {
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

	resp, err := kssh.GetSignedKey(config, shared.SignatureRequest{
		UUID:         randomUUID.String(),
		SSHPublicKey: string(pubKey),
	})
	if err != nil {
		return fmt.Errorf("Failed to get a signed key from the CA: %v", err)
	}

	err = ioutil.WriteFile(shared.KeyPathToCert(keyPath), []byte(resp.SignedKey), 0600)
	if err != nil {
		return fmt.Errorf("Failed to write new SSH key to disk: %v", err)
	}

	return nil
}

// Run SSH with the given key. Calls os.Exit if SSH returns
func runSSHWithKey(keyPath string, remainingArgs []string) {
	argumentList := []string{"-i", keyPath, "-o", "IdentitiesOnly=yes"}
	argumentList = append(argumentList, remainingArgs...)

	cmd := exec.Command("ssh", argumentList...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	err := cmd.Run()
	if err != nil {
		fmt.Printf("SSH exited with err: %v\n", err)
		os.Exit(1)
	}
	os.Exit(0)
}
