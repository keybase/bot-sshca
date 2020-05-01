package kssh

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/keybase/bot-sshca/src/shared"
)

// A Config that is provided by the keybaseca server process and lives in kbfs. It is used to share configuration
// information about how kssh should communicate with the keybaseca chatbot.
type Config struct {
	TeamName    string `json:"teamname"`
	ChannelName string `json:"channelname"`
	BotName     string `json:"botname"`
}

//TODO dedup
func (b *Bot) getAllTeams() (teams []string, err error) {
	memberships, err := b.api.ListUserMemberships(b.api.GetUsername())
	if err != nil {
		fmt.Printf("Failed to delete client configs: %v", err)
		return teams, err
	}
	for _, m := range memberships {
		teams = append(teams, m.FqName)
	}
	return teams, nil
}

// LoadConfigs loads client configs from KBFS. Returns a (listOfConfigs, listOfBotNames, err)
// Both lists are deduplicated based on Config.BotName. Runs the KBFS operations in parallel
// to speed up loading configs.
func (b *Bot) LoadConfigs() ([]Config, []string, error) {
	teams, err := b.getAllTeams()
	if err != nil {
		return nil, nil, err
	}
	botNameToConfig := make(map[string]Config)
	for _, team := range teams {
		conf, err := b.LoadConfig(team)
		if err != nil {
			return nil, nil, err
		}
		if conf != nil {
			botNameToConfig[conf.BotName] = *conf
		}
	}

	var configs []Config
	var botnames []string
	for _, config := range botNameToConfig {
		botnames = append(botnames, config.BotName)
		configs = append(configs, config)
	}

	return configs, botnames, nil
}

func (b *Bot) LoadConfig(teamName string) (*Config, error) {
	res, err := b.api.GetEntry(&teamName, shared.SSHCANamespace, shared.SSHCAConfigKey)
	if err == nil && (res.Revision > 0 && len(res.EntryValue) > 0) {
		var conf Config
		if err := json.Unmarshal([]byte(res.EntryValue), &conf); err != nil {
			return nil, fmt.Errorf("Failed to parse config for team %s: %v", teamName, err)
		}
		if conf.TeamName == "" || conf.BotName == "" {
			return nil, fmt.Errorf("found a config for team %s that is missing data: %s", teamName, res.EntryValue)
		}
		return &conf, nil
	}
	return nil, nil
}

// A LocalConfig is a file that lives on the FS of the computer running kssh. By default (and for most users), this
// file is not used.
//
// If a user of kssh is in in multiple teams that are running the CA bot they can configure a default bot to communicate
// with. Note that we store the team in here (even though it wasn't specified by the user) so that we can avoid doing
// a call to `LoadConfigs` if a default is set (since `LoadConfigs can be very slow if the user is in a large number of teams).
// This is controlled via `kssh --set-default-bot foo`.
//
// If a user of kssh wishes to configure a default ssh user to use (see README.md for a description of why this may
// be useful) this is also stored in the local config file. This is controlled via `kssh --set-default-user foo`.
type LocalConfig struct {
	DefaultBotName string `json:"default_bot"`
	DefaultBotTeam string `json:"default_team"`
	DefaultSSHUser string `json:"default_ssh_user"`
	KeybaseBinPath string `json:"keybase_binary"`
}

func GetKeybaseBinaryPath() string {
	lcf, err := getCurrentConfig()
	if err != nil {
		return "keybase"
	}

	if lcf.KeybaseBinPath != "" {
		return lcf.KeybaseBinPath
	}
	return "keybase"
}

// Where to store the local config file. Just stash it in ~/.ssh rather than making a ~/.kssh folder
var localConfigLocation = shared.ExpandPathWithTilde("~/.ssh/kssh-config.json")

// Get the default SSH user to use for kssh connections. Empty if no user is configured.
func GetDefaultSSHUser() (string, error) {
	lcf, err := getCurrentConfig()
	if err != nil {
		return "", err
	}

	return lcf.DefaultSSHUser, nil
}

// Set the default SSH user to use for kssh connections.
func SetKeybaseBinaryPath(path string) error {
	lcf, err := getCurrentConfig()
	if err != nil {
		return err
	}

	lcf.KeybaseBinPath = path
	return writeConfig(lcf)
}

// Set the default SSH user to use for kssh connections.
func SetDefaultSSHUser(username string) error {
	if strings.ContainsAny(username, " \t\n\r'\"") {
		return fmt.Errorf("invalid username: %s", username)
	}

	lcf, err := getCurrentConfig()
	if err != nil {
		return err
	}

	lcf.DefaultSSHUser = username
	return writeConfig(lcf)
}

// Write the given config file to disk
func writeConfig(lcf LocalConfig) error {
	bytes, err := json.Marshal(&lcf)
	if err != nil {
		return fmt.Errorf("failed to marshal json into config file: %v", err)
	}

	// Create ~/.ssh/ if it does not yet exist
	err = MakeDotSSH()
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(localConfigLocation, bytes, 0600)
	if err != nil {
		return fmt.Errorf("failed to write config file: %v", err)
	}
	return nil
}

// Get the current kssh config file
func getCurrentConfig() (lcf LocalConfig, err error) {
	if _, err := os.Stat(localConfigLocation); os.IsNotExist(err) {
		return lcf, nil
	}
	bytes, err := ioutil.ReadFile(localConfigLocation)
	if err != nil {
		return lcf, fmt.Errorf("failed to read local config file: %v", err)
	}
	err = json.Unmarshal(bytes, &lcf)
	if err != nil {
		return lcf, fmt.Errorf("failed to parse local config file: %v", err)
	}
	return lcf, nil
}

func ClearDefaultBot() error {
	return SetDefaultBot("")
}

// Set the default keybaseca bot to communicate with.
func SetDefaultBot(botName string) error {
	teamName := ""
	if botName != "" {
		bot, err := NewBot()
		if err != nil {
			return err
		}
		// Get the team associated with it and cache that too in order to avoid looking it up everytime
		conf, err := bot.LoadConfigForBot(botName)
		if err != nil {
			return err
		}
		teamName = conf.TeamName
	}

	lcf, err := getCurrentConfig()
	if err != nil {
		return err
	}
	lcf.DefaultBotName = botName
	lcf.DefaultBotTeam = teamName

	return writeConfig(lcf)
}

// Get the default bot and team for kssh
func GetDefaultBotAndTeam() (string, string, error) {
	lcf, err := getCurrentConfig()
	if err != nil {
		return "", "", err
	}
	return lcf.DefaultBotName, lcf.DefaultBotTeam, nil
}

// Get the teamname associated with the given botName
func (b *Bot) LoadConfigForBot(botName string) (Config, error) {
	configs, _, err := b.LoadConfigs()
	if err != nil {
		return Config{}, err
	}
	for _, config := range configs {
		if config.BotName == botName {
			return config, nil
		}
	}
	return Config{}, fmt.Errorf("did not find a client config file matching botName=%s (is the CA bot running and are you in the correct teams?)", botName)
}
