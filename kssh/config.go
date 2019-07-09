package kssh

import (
	"encoding/json"
	"io/ioutil"
	"path/filepath"

	"github.com/keybase/bot-ssh-ca/shared"
)

type ConfigFile struct {
	TeamName string `json:"teamname"`
	BotName  string `json:"botname"`
}

func LoadConfigs() ([]ConfigFile, error) {
	matches, _ := filepath.Glob("/keybase/team/*/" + shared.ConfigFilename)
	var ret []ConfigFile
	for _, match := range matches {
		conf, err := loadConfig(match)
		if err != nil {
			return nil, err
		}
		ret = append(ret, conf)
	}
	return ret, nil
}

func loadConfig(filename string) (ConfigFile, error) {
	var cf ConfigFile
	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return cf, err
	}
	err = json.Unmarshal(bytes, &cf)
	return cf, err
}
