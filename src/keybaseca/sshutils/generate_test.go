package sshutils

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/keybase/bot-ssh-ca/src/shared"
	"github.com/stretchr/testify/require"
)

func TestGenerateNewSSHKey(t *testing.T) {
	filename := "/tmp/bot-ssh-ca-integration-test-generate-key"
	os.Remove(filename)

	err := GenerateNewSSHKey(filename, false, false)
	require.NoError(t, err)

	err = GenerateNewSSHKey(filename, false, false)
	require.Errorf(t, err, "Refusing to overwrite existing key (try with --overwrite-existing-key or FORCE_WRITE=true if you're sure): "+filename)

	err = GenerateNewSSHKey(filename, true, false)
	require.NoError(t, err)

	bytes, err := ioutil.ReadFile(filename)
	require.NoError(t, err)
	require.True(t, strings.Contains(string(bytes), "PRIVATE"))

	bytes, err = ioutil.ReadFile(shared.KeyPathToPubKey(filename))
	require.NoError(t, err)
	require.False(t, strings.Contains(string(bytes), "PRIVATE"))
	require.True(t, strings.HasPrefix(string(bytes), "ssh-"))
}
