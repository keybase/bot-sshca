package botwrapper

// This package is meant to be a leaf node in the dependency graph for this project. It is required because the
// `bot` package and the `config` package both need a way of getting a KBChat API object and the `bot` package
// depends on the `config` package. Thus, this method could not live in the `bot` package without creating a
// dependency cycle.

import (
	"github.com/keybase/go-keybase-chat-bot/kbchat"
)

// Get a running instance of the keybase chat API. Will use the supplied credentials if necessary. If possible, it
// is preferred to reference the `GetKBChat` method in the `bot` package instead
func GetKBChat(keybaseHomeDir, keybasePaperKey, keybaseUsername string) (*kbchat.API, error) {
	runOptions := kbchat.RunOptions{}
	if keybaseHomeDir != "" {
		runOptions.HomeDir = keybaseHomeDir
	}
	if keybasePaperKey != "" && keybaseUsername != "" {
		runOptions.Oneshot = &kbchat.OneshotOptions{PaperKey: keybasePaperKey, Username: keybaseUsername}
	}
	return kbchat.Start(runOptions)
}
