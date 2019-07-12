package kssh

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/keybase/bot-ssh-ca/shared"
)

// A ConfigFile that is provided by the keybaseca server process and lives in kbfs
type ConfigFile struct {
	TeamName string `json:"teamname"`
	BotName  string `json:"botname"`
}

// LoadConfigs loads client configs from KBFS. Returns a (listOfConfigFiles, listOfTeamNames, err)
func LoadConfigs() ([]ConfigFile, []string, error) {
	matches, _ := filepath.Glob("/keybase/team/*/" + shared.ConfigFilename)
	var configs []ConfigFile
	var teams []string
	for _, match := range matches {
		conf, err := LoadConfig(match)
		if err != nil {
			return nil, nil, err
		}
		configs = append(configs, conf)
		teams = append(teams, strings.Split(match, "/")[3])
	}
	return configs, teams, nil
}

func LoadConfig(filename string) (ConfigFile, error) {
	var cf ConfigFile
	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return cf, err
	}
	err = json.Unmarshal(bytes, &cf)
	if cf.TeamName == "" || cf.BotName == "" {
		return cf, fmt.Errorf("Got a config file that is missing data: %s", string(bytes))
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
