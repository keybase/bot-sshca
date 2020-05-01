package constants

import "github.com/keybase/bot-sshca/src/keybaseca/kbfs"

// Get the default KBFSOperation struct used for KBFS operations. Currently
// keybaseca does not support running with a custom keybase binary path
func GetDefaultKBFSOperationsStruct() *kbfs.KBFSOperation {
	return &kbfs.KBFSOperation{KeybaseBinaryPath: "keybase"}
}
