package kssh

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"sync"

	"github.com/keybase/bot-ssh-ca/src/shared"
)

// A ConfigFile that is provided by the keybaseca server process and lives in kbfs. It is used to share configuration
// information about how kssh should communicate with the keybaseca chatbot.
type ConfigFile struct {
	TeamName    string `json:"teamname"`
	ChannelName string `json:"channelname"`
	BotName     string `json:"botname"`
}

// LoadConfigs loads client configs from KBFS. Returns a (listOfConfigFiles, listOfBotNames, err)
// Both lists are deduplicated based on ConfigFile.BotName. Runs the KBFS operations in parallel
// to speed up loading configs.
func LoadConfigs() ([]ConfigFile, []string, error) {
	allTeamsFromKBFS, err := shared.KBFSList("/keybase/team/")
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load config file(s): %v", err)
	}

	// Iterate through the listed files in parallel to speed up kssh for users with lots of teams
	semaphore := sync.WaitGroup{}
	semaphore.Add(len(allTeamsFromKBFS))
	boundChan := make(chan interface{}, shared.BoundedParallelismLimit)
	errors := make(chan error, len(allTeamsFromKBFS))
	botNameToConfig := make(map[string]ConfigFile)
	botNameToConfigMutex := sync.Mutex{}
	for _, team := range allTeamsFromKBFS {
		go func(team string) {
			// Blocks until there is room in boundChan
			boundChan <- 0

			filename := fmt.Sprintf("/keybase/team/%s/%s", team, shared.ConfigFilename)
			exists, err := shared.KBFSFileExists(filename)
			if err != nil {
				// Treat an error as it not existing and just skip that team while searching for config files
				exists = false
			}
			if exists {
				conf, err := LoadConfig(filename)
				if err != nil {
					errors <- err
				} else {
					botNameToConfigMutex.Lock()
					botNameToConfig[conf.BotName] = conf
					botNameToConfigMutex.Unlock()
				}
			}

			semaphore.Done()

			// Make room in boundChan
			<-boundChan
		}(team)
	}
	semaphore.Wait()

	// Read from errors without blocking
	select {
	case err := <-errors:
		return nil, nil, err
	default:
		// No error
	}

	var configs []ConfigFile
	var botnames []string
	for _, config := range botNameToConfig {
		botnames = append(botnames, config.BotName)
		configs = append(configs, config)
	}

	return configs, botnames, nil
}

// Load a kssh client config file from the given filename
func LoadConfig(kbfsFilename string) (ConfigFile, error) {
	var cf ConfigFile
	if !strings.HasPrefix(kbfsFilename, "/keybase/") {
		return cf, fmt.Errorf("cannot load a kssh config from outside of KBFS")
	}
	bytes, err := shared.KBFSRead(kbfsFilename)
	if err != nil {
		return cf, fmt.Errorf("found a config file at %s that could not be read: %v", kbfsFilename, err)
	}
	err = json.Unmarshal(bytes, &cf)
	if err != nil {
		return cf, fmt.Errorf("failed to parse config file at %s: %v", kbfsFilename, err)
	}
	if cf.TeamName == "" || cf.BotName == "" {
		return cf, fmt.Errorf("found a config file at %s that is missing data: %s", kbfsFilename, string(bytes))
	}
	return cf, err
}

// A LocalConfigFile is a file that lives on the FS of the computer running kssh. By default (and for most users), this
// file is not used.
//
// If a user of kssh is in in multiple teams that are running the CA bot they can configure a default bot to communicate
// with. Note that we store the team in here (even though it wasn't specified by the user) so that we can avoid doing
// a call to `LoadConfigs` if a default is set (since `LoadConfigs can be very slow if the user is in a large number of teams).
// This is controlled via `kssh --set-default-bot foo`.
//
// If a user of kssh wishes to configure a default ssh user to use (see README.md for a description of why this may
// be useful) this is also stored in the local config file. This is controlled via `kssh --set-default-user foo`.
type LocalConfigFile struct {
	DefaultBotName string `json:"default_bot"`
	DefaultBotTeam string `json:"default_team"`
	DefaultSSHUser string `json:"default_ssh_user"`
}

// Where to store the local config file. Just stash it in ~/.ssh rather than making a ~/.kssh folder
var localConfigFileLocation = shared.ExpandPathWithTilde("~/.ssh/kssh-config.json")

// Get the default SSH user to use for kssh connections. Empty if no user is configured.
func GetDefaultSSHUser() (string, error) {
	lcf, err := getCurrentConfig()
	if err != nil {
		return "", err
	}

	return lcf.DefaultSSHUser, nil
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
func writeConfig(lcf LocalConfigFile) error {
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
func getCurrentConfig() (lcf LocalConfigFile, err error) {
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

// Set the default keybaseca bot to communicate with.
func SetDefaultBot(botname string) error {
	teamname := ""
	var err error
	if botname != "" {
		// Get the team associated with it and cache that too in order to avoid looking it up everytime
		teamname, err = GetTeamFromBot(botname)
		if err != nil {
			return err
		}
	}

	lcf, err := getCurrentConfig()
	if err != nil {
		return err
	}
	lcf.DefaultBotName = botname
	lcf.DefaultBotTeam = teamname

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

// Get the teamname associated with the given botname
func GetTeamFromBot(botname string) (string, error) {
	configs, _, err := LoadConfigs()
	if err != nil {
		return "", err
	}
	for _, config := range configs {
		if config.BotName == botname {
			return config.TeamName, nil
		}
	}
	return "", fmt.Errorf("did not find a client config file matching botname=%s (is the CA bot running and are you in the correct teams?)", botname)
}
