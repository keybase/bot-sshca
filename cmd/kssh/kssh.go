package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"time"

	"github.com/google/uuid"
	"github.com/keybase/bot-ssh-ca/keybaseca/libca"
	"github.com/keybase/bot-ssh-ca/keybaseca/sshutils"
	"github.com/keybase/bot-ssh-ca/kssh"
	"github.com/keybase/bot-ssh-ca/shared"

	"golang.org/x/crypto/ssh"
)

func main() {
	keyPath := libca.ExpandPathWithTilde("~/.ssh/keybase-ca-key")
	if isValidCert(keyPath) {
		runSSHWithKey(keyPath)
	}
	configs, err := kssh.LoadConfigs()
	if err != nil {
		fmt.Printf("Failed to load config file(s): %v\n", err)
		return
	}
	if len(configs) == 1 {
		provisionNewKey(configs[0], keyPath)
		runSSHWithKey(keyPath)
	} else {
		// TODO: Not implemented yet
		panic("It is currently only supported to use kssh within one team!")
	}
}

func isValidCert(keyPath string) bool {
	_, err1 := os.Stat(keyPath)
	_, err2 := os.Stat(shared.KeyPathToPubKey(keyPath))
	_, err3 := os.Stat(shared.KeyPathToCert(keyPath))
	if os.IsNotExist(err1) || os.IsNotExist(err2) || os.IsNotExist(err3) {
		return false // Cert does not exist
	}

	certBytes, err := ioutil.ReadFile(shared.KeyPathToCert(keyPath))
	if err != nil {
		// Failed to read the file for some reason, just provision a new cert
		return false
	}
	k, _, _, _, err := ssh.ParseAuthorizedKey(certBytes)
	if err != nil {
		// Failed to parse it so just provision a new cert
		return false
	}
	// This is legal, see: https://github.com/golang/go/issues/22046
	cert := k.(*ssh.Certificate)
	validBefore := time.Unix(int64(cert.ValidBefore), 0)
	validAfter := time.Unix(int64(cert.ValidAfter), 0)
	return time.Now().After(validAfter) && time.Now().Before(validBefore)
}

func provisionNewKey(config kssh.ConfigFile, keyPath string) {
	err := sshutils.GenerateNewSSHKey(keyPath, true, false)
	if err != nil {
		fmt.Printf("Failed to generate a new SSH key: %v\n", err)
		return
	}
	pubKey, err := ioutil.ReadFile(shared.KeyPathToPubKey(keyPath))
	if err != nil {
		fmt.Printf("Failed to read the SSH key from the filesystem: %v\n", err)
		return
	}

	randomUUID, err := uuid.NewRandom()
	if err != nil {
		fmt.Printf("Failed to generate a new UUID for the SignatureRequest: %v\n", err)
		return
	}

	resp, err := kssh.GetSignedKey(config, shared.SignatureRequest{
		UUID:         randomUUID.String(),
		SSHPublicKey: string(pubKey),
	})
	if err != nil {
		fmt.Printf("Failed to get a signed key from the CA: %v\n", err)
		return
	}

	err = ioutil.WriteFile(shared.KeyPathToCert(keyPath), []byte(resp.SignedKey), 0600)
	if err != nil {
		fmt.Printf("Failed to write new SSH key to disk: %v\n", err)
		return
	}
}

func runSSHWithKey(keyPath string) {
	argumentList := []string{"-i", keyPath, "-o", "IdentitiesOnly=yes"}
	argumentList = append(argumentList, os.Args[1:]...)

	cmd := exec.Command("ssh", argumentList...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Run()
	os.Exit(0)
}
