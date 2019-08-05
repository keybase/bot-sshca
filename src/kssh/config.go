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

// LoadConfigs loads client configs from KBFS. Returns a (listOfConfigFiles, listOfBotNames, err)
// Both lists are deduplicated based on ConfigFile.BotName
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

// A LocalConfigFile is a file that lives on the FS of the computer running kssh. It is only used if the user is
// in multiple teams that are running the CA bot and they set a default bot via `kssh --set-default-bot foo`
// Note that we store the team in here (even though it wasn't specified by the user) so that we can avoid doing a
// a call to `LoadConfigs` if a default is set (since `LoadConfigs can be very slow if the user is in a large number
// of teams).
type LocalConfigFile struct {
	DefaultBotName string `json:"default_bot"`
	DefaultBotTeam string `json:"default_team"`
}

// Where to store the local config file. Just stash it in ~/.ssh
var localConfigFileLocation = shared.ExpandPathWithTilde("~/.ssh/kssh.config")

func SetDefaultBot(botname string) error {
	if botname == "" {
		return os.Remove(localConfigFileLocation)
	}
	teamname, err := GetTeamFromBot(botname)
	if err != nil {
		return err
	}
	bytes, err := json.Marshal(&LocalConfigFile{DefaultBotName: botname, DefaultBotTeam: teamname})
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(localConfigFileLocation, bytes, 0600)
	if err != nil {
		return err
	}
	return nil
}

func GetDefaultBotAndTeam() (string, string, error) {
	if _, err := os.Stat(localConfigFileLocation); os.IsNotExist(err) {
		return "", "", nil
	}
	bytes, err := ioutil.ReadFile(localConfigFileLocation)
	if err != nil {
		return "", "", err
	}
	var lcf LocalConfigFile
	err = json.Unmarshal(bytes, &lcf)
	if err != nil {
		return "", "", err
	}
	return lcf.DefaultBotName, lcf.DefaultBotTeam, nil
}

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
