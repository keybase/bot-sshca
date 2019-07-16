package kssh

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
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
	cmd := exec.Command("keybase", "fs", "ls", "-1", "--nocolor", "/keybase/team/")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to list files in /keybase/team/: %s (%v)", string(output), err)
	}

	var configs []ConfigFile
	var teams []string
	for _, team := range strings.Split(string(output), "\n") {
		if team == "" {
			continue
		}
		filename := fmt.Sprintf("/keybase/team/%s/%s", team, shared.ConfigFilename)
		exists, err := KBFSFileExists(filename)
		if err != nil {
			// Treat an error as it not existing and just skip that team while searching for config files
			exists = false
		}
		if exists {
			conf, err := LoadConfig(filename)
			if err != nil {
				return nil, nil, err
			}
			configs = append(configs, conf)
			teams = append(teams, team)
		}
	}
	return configs, teams, nil
}

func LoadConfig(kbfsFilename string) (ConfigFile, error) {
	var cf ConfigFile
	bytes, err := ReadKBFS(kbfsFilename)
	if err != nil {
		return cf, err
	}
	err = json.Unmarshal(bytes, &cf)
	if cf.TeamName == "" || cf.BotName == "" {
		return cf, fmt.Errorf("Got a config file that is missing data: %s", string(bytes))
	}
	return cf, err
}

func KBFSFileExists(kbfsFilename string) (bool, error) {
	cmd := exec.Command("keybase", "fs", "stat", kbfsFilename)
	bytes, err := cmd.CombinedOutput()
	if err == nil {
		return true, nil
	}
	if strings.Contains(string(bytes), "ERROR file does not exist") {
		return false, nil
	}
	return false, fmt.Errorf("failed to stat %s: %s (%v)", kbfsFilename, string(bytes), err)
}

func ReadKBFS(kbfsFilename string) ([]byte, error) {
	if !strings.HasPrefix(kbfsFilename, "/keybase/") {
		return nil, fmt.Errorf("cannot load a kssh config from outside of KBFS")
	}
	cmd := exec.Command("keybase", "fs", "read", kbfsFilename)
	bytes, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %s (%v)", kbfsFilename, string(bytes), err)
	}
	return bytes, nil
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
