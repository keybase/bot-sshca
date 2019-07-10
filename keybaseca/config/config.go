/*
A keybaseca config file looks like

```
# The location of the CA key file. Defaults to ~/keybaseca.config
ca_key_location: ~/keybase-ca-key
# How long signed keys are valid for. Defaults to 1 hour. Valid formats are +1h, +5h, +1d, +3d, +1w, etc
key_expiration: "+2h"
# The ssh user
user: root
# The name of the subteam used for granting SSH access
teamname: my_team.ssh

# Whether to use an alternate account. Only useful if you are running the chatbot on an account other than the one you are currently using
# Mainly useful for dev work
use_alternate_account: true
keybase_home_dir: /tmp/keybase/
keybase_paper_key: "paper key goes here"
keybase_username: username_for_the_bot
```
*/

package config

import (
	"io/ioutil"

	"github.com/keybase/bot-ssh-ca/keybaseca/libca"

	"github.com/go-yaml/yaml"
)

// Used by the CLI argument parsing code
var DefaultConfigLocation = libca.ExpandPathWithTilde("~/keybaseca.config")

// Represents a loaded config for keybaseca
type Config interface {
	GetCAKeyLocation() string
	GetKeybaseHomeDir() string
	GetKeybasePaperKey() string
	GetKeybaseUsername() string
	GetKeyExpiration() string
	GetSSHUser() string
	GetTeamName() string
}

// Load a yaml config file from the given filename. See the top of this file for an example yaml config file.
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
	CAKeyLocation   string `yaml:"ca_key_location"`
	KeybaseHomeDir  string `yaml:"keybase_home_dir"`
	KeybasePaperKey string `yaml:"keybase_paper_key"`
	KeybaseUsername string `yaml:"keybase_username"`
	KeyExpiration   string `yaml:"key_expiration"`
	SSHUser         string `yaml:"user"`
	TeamName        string `yaml:"teamname"`
}

var _ Config = (*ConfigFile)(nil)

func (cf *ConfigFile) GetCAKeyLocation() string {
	if cf.CAKeyLocation != "" {
		return libca.ExpandPathWithTilde(cf.CAKeyLocation)
	}
	return libca.ExpandPathWithTilde("~/keybase-ca-key")
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
