package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/keybase/bot-ssh-ca/src/shared"
	"github.com/stretchr/testify/require"
)

var certTestFilename = "/tmp/bot-ssh-ca-test-is-valid-cert"

func copyKey(t *testing.T, name string) {
	priv, err := ioutil.ReadFile(fmt.Sprintf("../../../tests/testFiles/%s", name))
	require.NoError(t, err)
	err = ioutil.WriteFile(certTestFilename, priv, 0600)
	require.NoError(t, err)
	pub, err := ioutil.ReadFile(fmt.Sprintf("../../../tests/testFiles/%s.pub", name))
	require.NoError(t, err)
	err = ioutil.WriteFile(shared.KeyPathToPubKey(certTestFilename), pub, 0600)
	require.NoError(t, err)
	cert, err := ioutil.ReadFile(fmt.Sprintf("../../../tests/testFiles/%s-cert.pub", name))
	require.NoError(t, err)
	err = ioutil.WriteFile(shared.KeyPathToCert(certTestFilename), cert, 0600)
	require.NoError(t, err)

}

func TestIsValidCert(t *testing.T) {
	os.Remove(certTestFilename)
	os.Remove(shared.KeyPathToPubKey(certTestFilename))
	os.Remove(shared.KeyPathToCert(certTestFilename))

	require.False(t, isValidCert(certTestFilename))

	copyKey(t, "valid")
	require.True(t, isValidCert(certTestFilename))

	copyKey(t, "expired")
	require.False(t, isValidCert(certTestFilename))
}

func BenchmarkLoadConfigs(b *testing.B) {
	os.Remove("~/.ssh/kssh.config")
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		getConfig("")
	}
}
