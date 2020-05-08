package kssh

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/keybase/bot-sshca/src/shared"
)

// Config is provided by the keybaseca server and lives in the KV store.  It is
// used to share configuration information about how kssh should communicate
// with the keybaseca chat bot.
type Config struct {
	TeamName    string `json:"teamname"`
	ChannelName string `json:"channelname"`
	BotName     string `json:"botname"`
}

// Get the configured channel name from the given config file. Returns either a pointer to the channel name string
// or a null pointer.
func (c *Config) getChannel() *string {
	if c.ChannelName != "" {
		return &c.ChannelName
	}
	return nil
}

// A LocalConfigFile is a file that lives on the FS of the computer running kssh.
// By default (and for most users), this file is not used.
//
// If a user of kssh is in in multiple teams that are running the CA bot they
// can configure a default bot to communicate with. Note that we store the team
// in here (even though it wasn't specified by the user) so that we can avoid
// doing a call to `LoadConfigs` if a default is set.  This is controlled via
// `kssh --set-default-bot foo`.
//
// If a user of kssh wishes to configure a default ssh user to use (see
// README.md for a description of why this may be useful) this is also stored
// in the local config file. This is controlled via `kssh --set-default-user
// foo`.
type LocalConfigFile struct {
	DefaultBotName string `json:"default_bot"`
	DefaultBotTeam string `json:"default_team"`
	DefaultSSHUser string `json:"default_ssh_user"`
	KeybaseBinPath string `json:"keybase_binary"`
}

func GetKeybaseBinaryPath() string {
	lcf, err := getCurrentConfigFile()
	if err != nil {
		return "keybase"
	}

	if lcf.KeybaseBinPath != "" {
		return lcf.KeybaseBinPath
	}
	return "keybase"
}

// Where to store the local config file. Just stash it in ~/.ssh rather than
// making a ~/.kssh folder
var localConfigFileLocation = shared.ExpandPathWithTilde("~/.ssh/kssh-config.json")

// Get the default SSH user to use for kssh connections. Empty if no user is configured.
func GetDefaultSSHUser() (string, error) {
	lcf, err := getCurrentConfigFile()
	if err != nil {
		return "", err
	}

	return lcf.DefaultSSHUser, nil
}

// Set the default SSH user to use for kssh connections.
func SetKeybaseBinaryPath(path string) error {
	lcf, err := getCurrentConfigFile()
	if err != nil {
		return err
	}

	lcf.KeybaseBinPath = path
	return writeConfigFile(lcf)
}

// Set the default SSH user to use for kssh connections.
func SetDefaultSSHUser(username string) error {
	if strings.ContainsAny(username, " \t\n\r'\"") {
		return fmt.Errorf("invalid username: %s", username)
	}

	lcf, err := getCurrentConfigFile()
	if err != nil {
		return err
	}

	lcf.DefaultSSHUser = username
	return writeConfigFile(lcf)
}

// Write the given config file to disk
func writeConfigFile(lcf LocalConfigFile) error {
	bytes, err := json.Marshal(&lcf)
	if err != nil {
		return fmt.Errorf("failed to marshal json into config file: %v", err)
	}

	// Create ~/.ssh/ if it does not yet exist
	err = MakeDotSSH()
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(localConfigFileLocation, bytes, 0600)
	if err != nil {
		return fmt.Errorf("failed to write config file: %v", err)
	}
	return nil
}

// Get the current kssh config file
func getCurrentConfigFile() (lcf LocalConfigFile, err error) {
	if _, err := os.Stat(localConfigFileLocation); os.IsNotExist(err) {
		return lcf, nil
	}
	bytes, err := ioutil.ReadFile(localConfigFileLocation)
	if err != nil {
		return lcf, fmt.Errorf("failed to read local config file: %v", err)
	}
	err = json.Unmarshal(bytes, &lcf)
	if err != nil {
		return lcf, fmt.Errorf("failed to parse local config file: %v", err)
	}
	return lcf, nil
}

// GetDefaultBotAndTeam gets the default bot and team for kssh from the local
// config file.
func GetDefaultBotAndTeam() (string, string, error) {
	lcf, err := getCurrentConfigFile()
	if err != nil {
		return "", "", err
	}
	return lcf.DefaultBotName, lcf.DefaultBotTeam, nil
}

// ClearDefaultBot clears the default bot and team.
func ClearDefaultBot() error {
	return SetDefaultBot("")
}

// SetDefaultBot sets the default keybaseca bot and team to communicate with.
// If given a botName, will need to start a Keybase bot to find and load all
// configs from KV store, in order to find the team associated with the given
// botName.
func SetDefaultBot(botName string) error {
	teamName := ""
	if botName != "" {
		requester, err := NewRequester()
		if err != nil {
			return err
		}
		// Get the team associated with it and cache that too in order to avoid
		// looking it up everytime
		conf, err := requester.LoadConfigForBot(botName)
		if err != nil {
			return err
		}
		teamName = conf.TeamName
	}

	lcf, err := getCurrentConfigFile()
	if err != nil {
		return err
	}
	lcf.DefaultBotName = botName
	lcf.DefaultBotTeam = teamName

	return writeConfigFile(lcf)
}
