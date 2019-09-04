package sshutils

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"golang.org/x/crypto/ssh"

	"github.com/keybase/bot-sshca/src/shared"
)

// Generate a new SSH key and store the private key at filename and the public key at filename.pub
// If the ssh-keygen binary exists, generates an ed25519 ssh key using ssh-keygen. Otherwise,
// generates an ecdsa key using go's crypto library. Note that we use ecdsa rather than ed25519
// in this case since go's crypto library does not support marshalling ed25519 keys into the format
// expected by openssh. github.com/ScaleFT/sshkeys claims to support this but does not reliably
// work with all versions of ssh.
func generateNewSSHKey(filename string) error {
	if sshKeygenBinaryExists() {
		return generateNewSSHKeyEd25519(filename)
	}

	return generateNewSSHKeyEcdsa(filename)
}

// Returns true iff the ssh-keygen binary exists and is in the user's path
func sshKeygenBinaryExists() bool {
	_, err := exec.LookPath("ssh-keygen")
	return err == nil
}

// Generate an ed25519 ssh key via ssh-keygen. Stores the private key at filename and the public key at filename.pub
func generateNewSSHKeyEd25519(filename string) error {
	cmd := exec.Command("ssh-keygen", "-t", "ed25519", "-f", filename, "-m", "PEM", "-N", "")
	bytes, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ssh-keygen failed: %s (%v)", strings.TrimSpace(string(bytes)), err)
	}
	return nil
}

// Generate an ecdsa ssh key in pure go code. Stores the private key at filename and the public key at filename.pub
// Note that if you are editing this code, be careful to ensure you test it manually since the integration tests
// run in an environment with ssh-keygen and thus do not call this function. This function is manually used on windows.
func generateNewSSHKeyEcdsa(filename string) error {
	// ssh-keygen -t ecdsa uses P256 by default
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return err
	}

	// 0600 are the correct permissions for an ssh private key
	privateKeyFile, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer privateKeyFile.Close()

	bytes, err := x509.MarshalECPrivateKey(privateKey)
	if err != nil {
		return err
	}

	privateKeyPEM := &pem.Block{Type: "EC PRIVATE KEY", Bytes: bytes}
	err = pem.Encode(privateKeyFile, privateKeyPEM)
	if err != nil {
		return err
	}

	pub, err := ssh.NewPublicKey(&privateKey.PublicKey)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(shared.KeyPathToPubKey(filename), ssh.MarshalAuthorizedKey(pub), 0600)
}
