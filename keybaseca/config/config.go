package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/keybase/bot-ssh-ca/shared"

	"github.com/go-yaml/yaml"
)

// Used by the CLI argument parsing code
var DefaultConfigLocation = shared.ExpandPathWithTilde("~/keybaseca.config")

// Represents a loaded config for keybaseca
type Config interface {
	GetCAKeyLocation() string
	GetKeybaseHomeDir() string
	GetKeybasePaperKey() string
	GetKeybaseUsername() string
	GetKeyExpiration() string
	GetSSHUser() string
	GetTeams() []string
	GetUseSubteamAsPrincipal() bool
	GetLogLocation() string
	GetStrictLogging() bool
}

// Load a yaml config file from the given filename. See the top of this file for an example yaml config file.
func LoadConfig(filename string) (Config, error) {
	contents, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	var cf ConfigFile
	// UnmarshalStrict will error out if there is an unexpected field in the yaml data
	err = yaml.UnmarshalStrict(contents, &cf)
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
	if cf.KeyExpiration != "" && !strings.HasPrefix(cf.KeyExpiration, "+") {
		// Only a basic check for this since ssh will error out later on if it is bogus
		return fmt.Errorf("key_expiration must be of the form `+<number><unit> where unit is one of `m`, `h`, `d`, `w`. Eg `+1h`. ")
	}
	if cf.LogLocation != "" && !isValidPath(cf.LogLocation) {
		return fmt.Errorf("log_location '%s' is not a valid path", cf.LogLocation)
	}
	return nil
}

func isValidPath(path string) bool {
	if strings.HasPrefix(path, "/keybase/") {
		// If it exists it is valid
		exists, err := shared.KBFSFileExists(path)
		if err != nil {
			return false
		}
		if exists {
			return true
		}

		// Otherwise try to write to it
		err = shared.KBFSWrite(path, "", false)
		if err != nil {
			return false
		}
		shared.KBFSDelete(path)
		return true
	}
	_, err := os.Stat(path)
	if err == nil {
		return true
	}

	var d []byte
	err = ioutil.WriteFile(path, d, 0600)
	if err != nil {
		return false
	}
	os.Remove(path)
	return true
}

type ConfigFile struct {
	CAKeyLocation         string   `yaml:"ca_key_location"`
	KeybaseHomeDir        string   `yaml:"keybase_home_dir"`
	KeybasePaperKey       string   `yaml:"keybase_paper_key"`
	KeybaseUsername       string   `yaml:"keybase_username"`
	KeyExpiration         string   `yaml:"key_expiration"`
	SSHUser               string   `yaml:"ssh_user"`
	Teams                 []string `yaml:"teams"`
	UseSubteamAsPrincipal bool     `yaml:"use_subteam_as_principal"`
	LogLocation           string   `yaml:"log_location"`
	StrictLogging         bool     `yaml:"strict_logging"`
}

var _ Config = (*ConfigFile)(nil)

func (cf *ConfigFile) GetCAKeyLocation() string {
	if cf.CAKeyLocation != "" {
		return shared.ExpandPathWithTilde(cf.CAKeyLocation)
	}
	return shared.ExpandPathWithTilde("~/keybase-ca-key")
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

func (cf *ConfigFile) GetUseSubteamAsPrincipal() bool {
	return cf.UseSubteamAsPrincipal
}

func (cf *ConfigFile) GetLogLocation() string {
	return cf.LogLocation
}

func (cf *ConfigFile) GetStrictLogging() bool {
	return cf.StrictLogging
}
