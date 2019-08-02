package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/keybase/bot-ssh-ca/src/shared"
	"github.com/stretchr/testify/require"
)

func copyKeyFromTextFixture(t *testing.T, name, destination string) {
	priv, err := ioutil.ReadFile(fmt.Sprintf("../../../tests/testFiles/%s", name))
	require.NoError(t, err)
	err = ioutil.WriteFile(destination, priv, 0600)
	require.NoError(t, err)
	pub, err := ioutil.ReadFile(fmt.Sprintf("../../../tests/testFiles/%s.pub", name))
	require.NoError(t, err)
	err = ioutil.WriteFile(shared.KeyPathToPubKey(destination), pub, 0600)
	require.NoError(t, err)
	cert, err := ioutil.ReadFile(fmt.Sprintf("../../../tests/testFiles/%s-cert.pub", name))
	require.NoError(t, err)
	err = ioutil.WriteFile(shared.KeyPathToCert(destination), cert, 0600)
	require.NoError(t, err)

}

func TestIsValidCert(t *testing.T) {
	certTestFilename := "/tmp/bot-ssh-ca-test-is-valid-cert"

	os.Remove(certTestFilename)
	os.Remove(shared.KeyPathToPubKey(certTestFilename))
	os.Remove(shared.KeyPathToCert(certTestFilename))

	require.False(t, isValidCert(certTestFilename))

	copyKeyFromTextFixture(t, "valid", certTestFilename)
	require.True(t, isValidCert(certTestFilename))

	copyKeyFromTextFixture(t, "expired", certTestFilename)
	require.False(t, isValidCert(certTestFilename))
}

func BenchmarkLoadConfigs(b *testing.B) {
	os.Remove("~/.ssh/kssh.config")
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		getConfig("")
	}
}
