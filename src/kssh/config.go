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

// LoadConfigs loads kssh configs from the KV store. Returns a (listOfConfigs,
// listOfBotNames, err). Both lists are deduplicated based on Config.BotName.
func (b *Bot) LoadConfigs() (configs []Config, botNames []string, err error) {
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
			// conf was found
			botNameToConfig[conf.BotName] = *conf
		}
	}
	for _, config := range botNameToConfig {
		configs = append(configs, config)
		botNames = append(botNames, config.BotName)
	}
	return configs, botNames, nil
}

// LoadConfig loads the kssh config for the given teamName. Will return a nil
// Config if no config was found for the teamName (and no error occurred)
func (b *Bot) LoadConfig(teamName string) (*Config, error) {
	res, err := b.api.GetEntry(&teamName, shared.SSHCANamespace, shared.SSHCAConfigKey)
	if err != nil {
		// error getting the entry
		return nil, err
	}
	if res.Revision > 0 && len(res.EntryValue) > 0 {
		// then this entry exists
		var conf Config
		if err := json.Unmarshal([]byte(res.EntryValue), &conf); err != nil {
			return nil, fmt.Errorf("Failed to parse config for team %s: %v", teamName, err)
		}
		if conf.TeamName == "" || conf.BotName == "" {
			return nil, fmt.Errorf("Found a config for team %s with missing data: %s", teamName, res.EntryValue)
		}
		return &conf, nil
	}
	// then entry doesn't exist; no config found
	return nil, nil
}

func (b *Bot) getAllTeams() (teams []string, err error) {
	// TODO: dedup with same method in keybaseca/bot
	memberships, err := b.api.ListUserMemberships(b.api.GetUsername())
	if err != nil {
		return teams, err
	}
	for _, m := range memberships {
		teams = append(teams, m.FqName)
	}
	return teams, nil
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

// LoadConfigForBot gets the Config associated with the given botName. Will
// need to find and load all configs from KV store.
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

	lcf, err := getCurrentConfigFile()
	if err != nil {
		return err
	}
	lcf.DefaultBotName = botName
	lcf.DefaultBotTeam = teamName

	return writeConfigFile(lcf)
}
