/*
A keybaseca config file looks like

```
# The location of the CA key file. Defaults to ~/keybaseca.config
ca_key_location: ~/keybase-ca-key
# How long signed keys are valid for. Defaults to 1 hour. Valid formats are +1h, +5h, +1d, +3d, +1w, etc
key_expiration: "+2h"
# The ssh user
ssh_user: root
# The name of the subteam used for granting SSH access
teams:
- my_team.ssh

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
	"fmt"
	"github.com/keybase/bot-ssh-ca/shared"
	"io/ioutil"
	"strings"

	"github.com/go-yaml/yaml"
)

// Used by the CLI argument parsing code
var DefaultConfigLocation = shared.ExpandPathWithTilde("~/keybaseca.config")

// Represents a loaded config for keybaseca
type Config interface {
	GetCAKeyLocation() string
	GetUseAlternateAccount() bool
	GetKeybaseHomeDir() string
	GetKeybasePaperKey() string
	GetKeybaseUsername() string
	GetKeyExpiration() string
	GetSSHUser() string
	GetTeams() []string
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
	err = validateConfig(cf)
	if err != nil {
		return nil, err
	}
	return &cf, nil
}

func validateConfig(cf ConfigFile) error {
	if len(cf.Teams) == 0 {
		return fmt.Errorf("must specify at least one team in the config file")
	}
	if cf.SSHUser == "" && cf.UseSubteamAsPrincipal == false {
		return fmt.Errorf("must specify either a ssh_user or use_subteam_as_principal")
	}
	if cf.SSHUser != "" && cf.UseSubteamAsPrincipal == true {
		return fmt.Errorf("cannot specify both a ssh_user and use_subteam_as_principal")
	}
	if cf.UseAlternateAccount && (cf.KeybaseHomeDir == "" || cf.KeybasePaperKey == "" || cf.KeybaseUsername == "") {
		return fmt.Errorf("Must specify keybase_home_dir, keybase_paper_key, and keybase_username if use_alternate_account is set")
	}
	if !cf.UseAlternateAccount && (cf.KeybaseHomeDir != "" || cf.KeybasePaperKey != "" || cf.KeybaseUsername != "") {
		return fmt.Errorf("keybase_home_dir, keybase_paper_key, and keybase_username cannot be set if use_alternate_account is not set")
	}
	if cf.KeyExpiration != "" && !strings.HasPrefix(cf.KeyExpiration, "+") {
		// Only a basic check for this since ssh will error out later on if it is bogus
		return fmt.Errorf("key_expiration must be of the form `+<number><unit> where unit is one of `m`, `h`, `d`, `w`. Eg `+1h`. ")
	}
	return nil
}

type ConfigFile struct {
	CAKeyLocation         string   `yaml:"ca_key_location"`
	UseAlternateAccount   bool     `yaml:"use_alternate_account"`
	KeybaseHomeDir        string   `yaml:"keybase_home_dir"`
	KeybasePaperKey       string   `yaml:"keybase_paper_key"`
	KeybaseUsername       string   `yaml:"keybase_username"`
	KeyExpiration         string   `yaml:"key_expiration"`
	SSHUser               string   `yaml:"ssh_user"`
	Teams                 []string `yaml:"teams"`
	UseSubteamAsPrincipal bool     `yaml:"use_subteam_as_principal"`
}

var _ Config = (*ConfigFile)(nil)

func (cf *ConfigFile) GetCAKeyLocation() string {
	if cf.CAKeyLocation != "" {
		return shared.ExpandPathWithTilde(cf.CAKeyLocation)
	}
	return shared.ExpandPathWithTilde("~/keybase-ca-key")
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

func (cf *ConfigFile) GetTeams() []string {
	return cf.Teams
}
