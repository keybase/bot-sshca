package constants

import "github.com/keybase/bot-sshca/src/shared"

// Get the default KBFSOperation struct used for KBFS operations. Currently keybaseca does not support running with a
// custom keybase binary path
func GetDefaultKBFSOperationsStruct() *shared.KBFSOperation {
	return &shared.KBFSOperation{KeybaseBinaryPath: "keybase"}
}
