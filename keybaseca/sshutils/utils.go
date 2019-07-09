package sshutils

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"

	"github.com/keybase/bot-ssh-ca/keybaseca/config"
	"github.com/keybase/bot-ssh-ca/shared"

	"github.com/google/uuid"
)

func GenerateNewSSHKey(filename string, overwrite bool, printPubKey bool) error {
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

	cmd := exec.Command("ssh-keygen", "-t", "ecdsa", "-f", filename, "-m", "PEM", "-N", "")
	err := cmd.Run()
	if err != nil {
		return err
	}
	if printPubKey {
		bytes, err := ioutil.ReadFile(shared.KeyPathToPubKey(filename))
		if err != nil {
			return err
		}
		fmt.Printf("Generated new public key: \n%s\n", string(bytes))
	}
	return nil
}

func Generate(conf config.Config, overwrite bool, printPubKey bool) error {
	err := GenerateNewSSHKey(conf.GetCAKeyLocation(), overwrite, printPubKey)
	if err == nil {
		fmt.Printf("Wrote new SSH CA key to %s\n", conf.GetCAKeyLocation())
	}
	return err
}

func getTempFilename(pattern string) (string, error) {
	f, err := ioutil.TempFile("", pattern)
	if err != nil {
		return "", err
	}
	tempFilename := f.Name()
	f.Close()
	err = os.Remove(tempFilename)
	if err != nil {
		return "", err
	}
	return tempFilename, nil
}

func ProcessSignatureRequest(conf config.Config, sr shared.SignatureRequest) (resp shared.SignatureResponse, err error) {
	randomUUID, err := uuid.NewRandom()
	if err != nil {
		return
	}
	principals, err := getPrincipals(conf, sr)
	if err != nil {
		return
	}

	tempFilename, err := getTempFilename("keybase-ca-signed-key")
	if err != nil {
		return
	}
	err = ioutil.WriteFile(shared.KeyPathToPubKey(tempFilename), []byte(sr.SSHPublicKey), 0600)
	if err != nil {
		return
	}

	cmd := exec.Command("ssh-keygen",
		"-s", conf.GetCAKeyLocation(), // The CA key
		"-I", sr.UUID+":"+randomUUID.String(), // The ID of the signed key. Use their uuid and our uuid to ensure it is unique
		"-n", principals, // The allowed principals
		"-V", conf.GetKeyExpiration(), // The configured key expiration
		"-N", "", // No password on the key
		shared.KeyPathToPubKey(tempFilename), // The location of where to put the key
	)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	err = cmd.Run()
	if err != nil {
		return
	}

	data, err := ioutil.ReadFile(shared.KeyPathToCert(tempFilename))
	if err != nil {
		return
	}
	err = os.Remove(shared.KeyPathToPubKey(tempFilename))
	if err != nil {
		return
	}
	err = os.Remove(shared.KeyPathToCert(tempFilename))
	if err != nil {
		return
	}
	return shared.SignatureResponse{SignedKey: string(data), UUID: sr.UUID}, nil
}

func getPrincipals(conf config.Config, sr shared.SignatureRequest) (string, error) {
	return conf.GetSSHUser(), nil
}
