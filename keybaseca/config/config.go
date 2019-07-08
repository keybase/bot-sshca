package config

import (
	"github.com/keybase/bot-ssh-ca/keybaseca/libca"
	"io/ioutil"

	"github.com/go-yaml/yaml"
)

type Config interface {
	GetCAKeyLocation() string
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
	CAKeyLocation string `yaml:"ca_key_location"`
}

var _ Config = (*ConfigFile)(nil)

func (cf *ConfigFile) GetCAKeyLocation() string {
	if cf.CAKeyLocation != "" {
		return cf.CAKeyLocation
	}
	return libca.ExpandPathWithTilde("~/keybase-ca-key")
}
