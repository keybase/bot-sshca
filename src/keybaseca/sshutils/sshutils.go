package sshutils

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"github.com/keybase/bot-ssh-ca/src/keybaseca/botwrapper"

	"github.com/keybase/bot-ssh-ca/src/keybaseca/log"

	"github.com/keybase/bot-ssh-ca/src/keybaseca/config"
	"github.com/keybase/bot-ssh-ca/src/shared"

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

// Get the principals that should be placed in the signed certificate.
// Note that this function is a security boundary since if it was bypassed an
// attacker would be able to provision SSH keys for environments that they should not have access to.
func getPrincipals(conf config.Config, sr shared.SignatureRequest) (string, error) {
	// Start by getting the list of teams the user is in
	api, err := botwrapper.GetKBChat(conf.GetKeybaseHomeDir(), conf.GetKeybasePaperKey(), conf.GetKeybaseUsername())
	if err != nil {
		return "", fmt.Errorf("failed to retrieve the list of teams the user is in: %v", err)
	}
	results, err := api.ListUserMemberships(sr.Username)
	if err != nil {
		return "", fmt.Errorf("failed to retrieve the list of teams the user is in: %v", err)
	}

	// Maps from a team to whether or not the user is in the current team (with writer, admin, or owner permissions)
	teamToMembership := make(map[string]bool)
	for _, result := range results {
		// Sadly result.Role is an integer and this is all we're given. Let's hope no one ever changes this enum out
		// from underneath us. Admittedly, the worst that could (should) happen is that someone with minimal permissions
		// in a team is given access (eg a reader) which wouldn't lead to a complete compromise since an attacker
		// would still have to be added as a reader first.
		//
		// result.Role == 4 --> owner
		// result.Role == 3 --> admin
		// result.Role == 2 --> writer
		if result.Role == 4 || result.Role == 3 || result.Role == 2 {
			teamToMembership[result.TeamName] = true
		}
	}

	// Iterate through the teams in the config file and use the subteam as the principal
	// if the user is in that subteam
	var principals []string
	for _, team := range conf.GetTeams() {
		result, ok := teamToMembership[team]
		if ok && result {
			principals = append(principals, team)
		}
	}
	return strings.Join(principals, ","), nil
}
