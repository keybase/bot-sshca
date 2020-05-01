package constants

import "github.com/keybase/bot-sshca/src/keybaseca/kbfs"

// Get the default KBFS Operation struct used for KBFS operations. Currently
// keybaseca does not support running with a custom keybase binary path
func GetDefaultKBFSOperationsStruct() *kbfs.Operation {
	return &kbfs.Operation{KeybaseBinaryPath: "keybase"}
}
