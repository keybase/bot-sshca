package generate

import (
	"fmt"
	"github.com/keybase/bot-ssh-ca/keybaseca/config"
	"os"
	"os/exec"
)

func generateNewSSHKey(filename string, overwrite bool) error {
	if _, err := os.Stat(filename); err == nil {
		if overwrite {
			err := os.Remove(filename)
			if err != nil {
				return err
			}
		} else {
			return fmt.Errorf("Refusing to overwrite existing key (try with --overwrite-existing-key if you're sure): %s", filename)
		}
	}

	cmd := exec.Command("ssh-keygen", "-t", "ed25519", "-f", filename, "-m", "PEM", "-N", "")

	return cmd.Run()
}

func Generate(conf config.Config, overwrite bool) error {
	err := generateNewSSHKey(conf.GetCAKeyLocation(), overwrite)
	if err == nil {
		fmt.Printf("Wrote new SSH CA key to %s\n", conf.GetCAKeyLocation())
	}
	return err
}
