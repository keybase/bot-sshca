package log

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/keybase/bot-ssh-ca/src/keybaseca/constants"

	"github.com/keybase/bot-ssh-ca/src/keybaseca/config"
)

// Log attempts to log the given string to a file. If conf.GetStrictLogging() it will panic if it fails
// to log to the file. If conf.GetStrictLogging() is false, it may silently fail
func Log(conf config.Config, str string) {
	strWithTs := fmt.Sprintf("[%s] %s", time.Now().String(), str)

	if conf.GetLogLocation() == "" {
		fmt.Print(strWithTs + "\n")
	} else {
		err := appendToFile(conf.GetLogLocation(), strWithTs)
		if err != nil {
			if conf.GetStrictLogging() {
				panic(fmt.Errorf("Failed to log '%s' to %s: %v", strings.TrimSpace(strWithTs), conf.GetLogLocation(), err))
			} else {
				fmt.Printf("Failed to log '%s' to %s: %v\n", strings.TrimSpace(strWithTs), conf.GetLogLocation(), err)
			}
		}
	}
}

// Append to the file at the given filename via either Keybase simple fs commands or via standard interactions with
// the local filesystem
func appendToFile(filename string, str string) error {
	if strings.HasPrefix(filename, "/keybase/") {
		return constants.GetDefaultKBFSOperationsStruct().KBFSWrite(filename, str, true)
	}
	f, err := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		return err
	}

	defer f.Close()
	_, err = f.WriteString(str)

	if err != nil {
		return err
	}

	return nil
}
