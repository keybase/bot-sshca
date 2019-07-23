package sshutils

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"github.com/keybase/bot-ssh-ca/keybaseca/log"

	"github.com/keybase/bot-ssh-ca/keybaseca/config"
	"github.com/keybase/bot-ssh-ca/shared"

	"github.com/google/uuid"
)

func GenerateNewSSHKey(filename string, overwrite bool, printPubKey bool) error {
	_, err := os.Stat(filename)
	if err == nil {
		if overwrite {
			err := os.Remove(filename)
			if err != nil {
				return err
			}
		} else {
			return fmt.Errorf("Refusing to overwrite existing key (try with --overwrite-existing-key or FORCE_WRITE=true if you're sure): %s", filename)
		}
	}

	err = generateNewSSHKey(filename)
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
		log.Log(conf, fmt.Sprintf("Wrote new SSH CA key to %s", conf.GetCAKeyLocation()))
	}
	return err
}

// Get a temporary filename that starts with pattern using ioutil.TempFile
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

	keyID := sr.UUID + ":" + randomUUID.String()

	log.Log(conf, fmt.Sprintf("Processing SignatureRequest from user=%s on device='%s' keyID:%s, principals:%s, expiration:%s, pubkey:%s",
		sr.Username, sr.DeviceName, keyID, principals, conf.GetKeyExpiration(), sr.SSHPublicKey))
	cmd := exec.Command("ssh-keygen",
		"-s", conf.GetCAKeyLocation(), // The CA key
		"-I", keyID, // The ID of the signed key. Use their uuid and our uuid to ensure it is unique
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

// Get the principals that should be placed in the signed certificate
func getPrincipals(conf config.Config, sr shared.SignatureRequest) (string, error) {
	if conf.GetUseSubteamAsPrincipal() {
		// Iterate through the teams in the config file and use the last portion of the subteam as the principal
		// if the user is in that subteam
		var principals []string
		for _, team := range conf.GetTeams() {
			members, err := getMembers(team)
			if err != nil {
				return "", err
			}
			for _, member := range members {
				if member == sr.Username {
					subteamChunks := strings.Split(team, ".")
					principals = append(principals, subteamChunks[len(subteamChunks)-1])
				}
			}
		}
		return strings.Join(principals, ","), nil
	}
	return conf.GetSSHUser(), nil
}

// Get the members of the given team
func getMembers(team string) ([]string, error) {
	cmd := exec.Command("keybase", "team", "list-members", team)
	data, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}
	var users []string
	for _, line := range strings.Split(string(data), "\n") {
		if strings.Contains(line, "writer") || strings.Contains(line, "admin") {
			users = append(users, strings.Fields(line)[2])
		}
	}
	return users, nil
}
