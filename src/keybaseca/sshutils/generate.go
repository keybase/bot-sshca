package sshutils

import (
	"crypto/rand"
	"fmt"
	"io/ioutil"

	"github.com/ScaleFT/sshkeys"
	"github.com/keybase/bot-sshca/src/shared"
	"golang.org/x/crypto/ed25519"
	"golang.org/x/crypto/ssh"
)

// Generate a new SSH key. Places the private key at filename and the public key at filename.pub.
// We use ed25519 keys since they may be more secure (and are smaller). The go crypto ssh library
// does not support marshalling ed25519 keys so we use ScaleFT/sshkeys to marshal them to the
// correct on disk format for SSH
func generateNewSSHKey(filename string) error {
	// Generate the key
	pub, private, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return fmt.Errorf("failed to generate ed25519 key: %v", err)
	}

	// Write the private key
	bytes, err := sshkeys.Marshal(private, &sshkeys.MarshalOptions{Format: sshkeys.FormatOpenSSHv1})
	if err != nil {
		return fmt.Errorf("failed to marshal ed25519 key: %v", err)
	}
	err = ioutil.WriteFile(filename, bytes, 0600)
	if err != nil {
		return fmt.Errorf("failed to write ssh private key to %s: %v", filename, err)
	}

	// Write the public key
	publicKey, err := ssh.NewPublicKey(pub)
	if err != nil {
		return fmt.Errorf("failed to create public key from ed25519 key: %v", err)
	}
	bytes = ssh.MarshalAuthorizedKey(publicKey)
	err = ioutil.WriteFile(shared.KeyPathToPubKey(filename), bytes, 0600)
	if err != nil {
		return fmt.Errorf("failed to write ssh public key to %s: %v", shared.KeyPathToPubKey(filename), err)
	}

	return nil
}
