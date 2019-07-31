package kssh

import (
	"fmt"
	"os/exec"
)

func AddKeyToSSHAgent(keyPath string) error {
	cmd := exec.Command("ssh-add", keyPath)
	bytes, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to add SSH key to the ssh-agent (is it running?): %s (%v)", string(bytes), err)
	}
	return nil
}
