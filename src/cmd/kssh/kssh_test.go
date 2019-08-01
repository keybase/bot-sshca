package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/keybase/bot-ssh-ca/src/shared"
	"github.com/stretchr/testify/assert"
)

var certTestFilename = "/tmp/bot-ssh-ca-test-is-valid-cert"

func copyKey(t *testing.T, name string) {
	priv, err := ioutil.ReadFile(fmt.Sprintf("../../../tests/testFiles/%s", name))
	assert.NoError(t, err)
	err = ioutil.WriteFile(certTestFilename, priv, 0600)
	assert.NoError(t, err)
	pub, err := ioutil.ReadFile(fmt.Sprintf("../../../tests/testFiles/%s.pub", name))
	assert.NoError(t, err)
	err = ioutil.WriteFile(shared.KeyPathToPubKey(certTestFilename), pub, 0600)
	assert.NoError(t, err)
	cert, err := ioutil.ReadFile(fmt.Sprintf("../../../tests/testFiles/%s-cert.pub", name))
	assert.NoError(t, err)
	err = ioutil.WriteFile(shared.KeyPathToCert(certTestFilename), cert, 0600)
	assert.NoError(t, err)

}

func TestIsValidCert(t *testing.T) {
	os.Remove(certTestFilename)
	os.Remove(shared.KeyPathToPubKey(certTestFilename))
	os.Remove(shared.KeyPathToCert(certTestFilename))

	assert.False(t, isValidCert(certTestFilename))

	copyKey(t, "valid")
	assert.True(t, isValidCert(certTestFilename))

	copyKey(t, "expired")
	assert.False(t, isValidCert(certTestFilename))
}
