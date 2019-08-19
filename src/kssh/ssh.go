package kssh

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/keybase/bot-ssh-ca/src/shared"
)

func AddKeyToSSHAgent(keyPath string) error {
	cmd := exec.Command("ssh-add", keyPath)
	bytes, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to add SSH key to the ssh-agent (is it running?): %s (%v)", strings.TrimSpace(string(bytes)), err)
	}
	return nil
}

var AlternateSSHConfigFile = shared.ExpandPathWithTilde("~/.ssh/kssh-config")

// Create an SSH config file that inherits from the default SSH config file but sets a default SSH user
func CreateDefaultUserConfigFile() error {
	user, err := GetDefaultSSHUser()
	if err != nil {
		return err
	}
	if user == "" {
		return nil
	}

	err = MakeDotSSH()
	if err != nil {
		return err
	}

	if _, err := os.Stat(shared.ExpandPathWithTilde("~/.ssh/config")); os.IsNotExist(err) {
		f, err := os.OpenFile(shared.ExpandPathWithTilde("~/.ssh/config"), os.O_RDONLY|os.O_CREATE, 0644)
		if err != nil {
			return fmt.Errorf("failed to touch ~/.ssh/config: %v", err)
		}
		f.Close()
	}

	config := fmt.Sprintf("# kssh config file to set a default SSH user\n"+
		"Include config\n"+
		"Host *\n"+
		"  User %s\n", user)

	f, err := os.OpenFile(AlternateSSHConfigFile, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString(config)
	if err != nil {
		return err
	}
	fmt.Printf("Using default ssh user %s\n", user)
	return nil
}

func MakeDotSSH() error {
	if _, err := os.Stat(shared.ExpandPathWithTilde("~/.ssh/")); os.IsNotExist(err) {
		err = os.Mkdir(shared.ExpandPathWithTilde("~/.ssh/"), 0700)
		if err != nil {
			return fmt.Errorf("failed to create ~/.ssh directory: %v", err)
		}
	}
	return nil
}
