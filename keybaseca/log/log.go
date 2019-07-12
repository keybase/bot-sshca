package log

import (
	"fmt"
	"os"
	"strings"

	"github.com/keybase/bot-ssh-ca/keybaseca/config"
	"github.com/keybase/bot-ssh-ca/shared"
)

// Log attempts to log the given string to a file. If conf.GetStrictLogging() it will panic if it fails
// to log to the file. If conf.GetStrictLogging() is false, it may silently fail
func Log(conf config.Config, str string) {
	if conf.GetLogLocation() == "" {
		fmt.Println(str)
	} else {
		err := appendToFile(conf.GetLogLocation(), str+"\n")
		if err != nil {
			if conf.GetStrictLogging() {
				panic(fmt.Errorf("Failed to log '%s' to %s: %v", str, conf.GetLogLocation(), err))
			} else {
				fmt.Printf("Failed to log '%s' to %s: %v\n", str, conf.GetLogLocation(), err)
			}
		}
	}
}

func appendToFile(filename string, str string) error {
	if strings.HasPrefix(filename, "/keybase/") {
		return shared.KBFSWrite(filename, str, true)
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
