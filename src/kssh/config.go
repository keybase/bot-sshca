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

// A ConfigFile that is provided by the keybaseca server process and lives in kbfs
type ConfigFile struct {
	TeamName    string `json:"teamname"`
	ChannelName string `json:"channelname"`
	BotName     string `json:"botname"`
}

// LoadConfigs loads client configs from KBFS. Returns a (listOfConfigFiles, listOfTeamNames, err)
// Both lists are deduplicated based on ConfigFile.BotName
func LoadConfigs() ([]ConfigFile, []string, error) {
	listedFiles, err := shared.KBFSList("/keybase/team/")
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load config file(s): %v", err)
	}

	// Iterate through the listed files in parallel to speed up kssh for users with lots of teams
	semaphore := make(chan interface{}, len(listedFiles))
	errors := make(chan error, len(listedFiles))
	botNameToConfig := make(map[string]ConfigFile)
	botNameToConfigMutex := sync.Mutex{}
	for _, team := range listedFiles {
		go func(team string) {
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

			semaphore <- 0
		}(team)
	}
	for i := 0; i < len(listedFiles); i++ {
		<-semaphore
	}

	// Read from errors without blocking
	select {
	case err := <-errors:
		return nil, nil, err
	default:
		// No error
	}

	var configs []ConfigFile
	var teams []string
	for _, config := range botNameToConfig {
		teams = append(teams, config.TeamName)
		configs = append(configs, config)
	}

	return configs, teams, nil
}

func LoadConfig(kbfsFilename string) (ConfigFile, error) {
	var cf ConfigFile
	if !strings.HasPrefix(kbfsFilename, "/keybase/") {
		return cf, fmt.Errorf("cannot load a kssh config from outside of KBFS")
	}
	bytes, err := shared.KBFSRead(kbfsFilename)
	if err != nil {
		return cf, err
	}
	err = json.Unmarshal(bytes, &cf)
	if cf.TeamName == "" || cf.BotName == "" {
		return cf, fmt.Errorf("Found a config file at %s that is missing data: %s", kbfsFilename, string(bytes))
	}
	return cf, err
}

// A LocalConfigFile is a file that lives on the FS of the computer running kssh. It is only used if the user is
// in multiple teams that are running the CA bot and they set a default team via `kssh --set-default-team foo`
type LocalConfigFile struct {
	DefaultTeam string `json:"default_team"`
}

// Where to store the local config file. Just stash it in ~/.ssh
var localConfigFileLocation = shared.ExpandPathWithTilde("~/.ssh/kssh.config")

func SetDefaultTeam(team string) error {
	bytes, err := json.Marshal(&LocalConfigFile{DefaultTeam: team})
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(localConfigFileLocation, bytes, 0600)
	if err != nil {
		return err
	}
	return nil
}

func GetDefaultTeam() (string, error) {
	if _, err := os.Stat(localConfigFileLocation); os.IsNotExist(err) {
		return "", nil
	}
	bytes, err := ioutil.ReadFile(localConfigFileLocation)
	if err != nil {
		return "", err
	}
	var lcf LocalConfigFile
	err = json.Unmarshal(bytes, &lcf)
	if err != nil {
		return "", err
	}
	return lcf.DefaultTeam, nil
}
