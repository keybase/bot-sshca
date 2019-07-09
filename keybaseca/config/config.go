package config

import (
	"io/ioutil"

	"github.com/keybase/bot-ssh-ca/keybaseca/libca"

	"github.com/go-yaml/yaml"
)

type Config interface {
	GetCAKeyLocation() string
	GetUseAlternateAccount() bool
	GetKeybaseHomeDir() string
	GetKeybasePaperKey() string
	GetKeybaseUsername() string
	GetKeyExpiration() string
	GetSSHUser() string
	GetTeamName() string
}

func LoadConfig(filename string) (Config, error) {
	contents, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	var cf ConfigFile
	err = yaml.Unmarshal(contents, &cf)
	if err != nil {
		return nil, err
	}
	return &cf, nil
}

type ConfigFile struct {
	CAKeyLocation       string `yaml:"ca_key_location"`
	UseAlternateAccount bool   `yaml:"use_alternate_account"`
	KeybaseHomeDir      string `yaml:"keybase_home_dir"`
	KeybasePaperKey     string `yaml:"keybase_paper_key"`
	KeybaseUsername     string `yaml:"keybase_username"`
	KeyExpiration       string `yaml:"key_expiration"`
	SSHUser             string `yaml:"user"`
	TeamName            string `yaml:"teamname"`
}

var _ Config = (*ConfigFile)(nil)

func (cf *ConfigFile) GetCAKeyLocation() string {
	if cf.CAKeyLocation != "" {
		return libca.ExpandPathWithTilde(cf.CAKeyLocation)
	}
	return libca.ExpandPathWithTilde("~/keybase-ca-key")
}

func (cf *ConfigFile) GetUseAlternateAccount() bool {
	return cf.UseAlternateAccount
}

func (cf *ConfigFile) GetKeybaseHomeDir() string {
	return cf.KeybaseHomeDir
}

func (cf *ConfigFile) GetKeybasePaperKey() string {
	return cf.KeybasePaperKey
}

func (cf *ConfigFile) GetKeybaseUsername() string {
	return cf.KeybaseUsername
}

func (cf *ConfigFile) GetKeyExpiration() string {
	if cf.KeyExpiration != "" {
		return cf.KeyExpiration
	}
	return "+1h"
}

func (cf *ConfigFile) GetSSHUser() string {
	return cf.SSHUser
}

func (cf *ConfigFile) GetTeamName() string {
	return cf.TeamName
}
