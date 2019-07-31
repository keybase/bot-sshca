// +build windows
// If you edit this file, be sure to test it on windows also. Our current test suite does not test windows support.

package sshutils

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"io/ioutil"
	"os"

	"github.com/keybase/bot-ssh-ca/src/shared"
	"golang.org/x/crypto/ssh"
)

// Generate a new SSH key. Places the private key at filename and the public key at filename.pub. If `overwrite`,
// it will overwrite the existing key. If `printPubKey` it will print out the generated public key to stdout.
// On windows, we use 2048 bit rsa keys. go's ssh library doesn't support ed25519 and ssh-keygen isn't built in.
func generateNewSSHKey(filename string) error {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return err
	}

	privateKeyFile, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer privateKeyFile.Close()

	privateKeyPEM := &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privateKey)}
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
