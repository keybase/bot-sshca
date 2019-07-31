// +build !windows

package sshutils

import (
	"fmt"
	"os/exec"
)

// Generate a new SSH key. Places the private key at filename and the public key at filename.pub. If `overwrite`,
// it will overwrite the existing key. If `printPubKey` it will print out the generated public key to stdout.
// On unix, we use ed25519 keys since they may be more secure (and are smaller).
func generateNewSSHKey(filename string) error {
	cmd := exec.Command("ssh-keygen", "-t", "ed25519", "-f", filename, "-m", "PEM", "-N", "")
	bytes, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ssh-keygen failed: %s (%v)", string(bytes), err)
	}
	return nil
}
