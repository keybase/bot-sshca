package shared

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
)

// Returns whether or not the current system supports accessing KBFS via a FUSE filesystem mounted at /keybase
// This is used in order to optimize heavily used functions in the below library. Generally, it is preferred to
// rely on `keybase fs` commands since those are guaranteed to work across systems (and are what is used inside the
// integration tests). But in a few cases (namely when kssh is searching for kssh-client.config files) it gives very
// large speed improvements to use the FUSE filesystem when available (an order of magnitude improvement for kssh)
func supportsFuse() bool {
	// Note that this function is not tested via integration tests since fuse does not run in docker. Handle with care.
	_, err1 := os.Stat("/keybase")
	_, err2 := os.Stat("/keybase/team")
	_, err3 := os.Stat("/keybase/private")
	_, err4 := os.Stat("/keybase/public")
	return err1 == nil && err2 == nil && err3 == nil && err4 == nil
}

func KBFSFileExists(kbfsFilename string) (bool, error) {
	if supportsFuse() {
		// Note that this code is not tested via integration tests since fuse does not run in docker. Handle with care.
		_, err := os.Stat(kbfsFilename)
		if err == nil {
			return true, nil
		}
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}

	cmd := exec.Command("keybase", "fs", "stat", kbfsFilename)
	bytes, err := cmd.CombinedOutput()
	if err == nil {
		return true, nil
	}
	if strings.Contains(string(bytes), "ERROR file does not exist") {
		return false, nil
	}
	return false, fmt.Errorf("failed to stat %s: %s (%v)", kbfsFilename, strings.TrimSpace(string(bytes)), err)
}

func KBFSRead(kbfsFilename string) ([]byte, error) {
	if supportsFuse() {
		// Note that this code is not tested via integration tests since fuse does not run in docker. Handle with care.
		return ioutil.ReadFile(kbfsFilename)
	}
	cmd := exec.Command("keybase", "fs", "read", kbfsFilename)
	bytes, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %s (%v)", kbfsFilename, strings.TrimSpace(string(bytes)), err)
	}
	return bytes, nil
}

func KBFSDelete(filename string) error {
	cmd := exec.Command("keybase", "fs", "rm", filename)
	bytes, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to delete the file at %s: %s (%v)", filename, strings.TrimSpace(string(bytes)), err)
	}
	return nil
}

func KBFSWrite(filename string, contents string, appendToFile bool) error {
	var cmd *exec.Cmd
	if appendToFile {
		// `keybase fs write --append` only works if the file already exists so create it if it does not exist
		exists, err := KBFSFileExists(filename)
		if !exists || err != nil {
			err = KBFSWrite(filename, "", false)
			if err != nil {
				return err
			}
		}
		cmd = exec.Command("keybase", "fs", "write", "--append", filename)
	} else {
		cmd = exec.Command("keybase", "fs", "write", filename)
	}

	cmd.Stdin = strings.NewReader(string(contents))
	bytes, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to write to file at %s: %s (%v)", filename, strings.TrimSpace(string(bytes)), err)
	}
	return nil
}

func KBFSList(path string) ([]string, error) {
	cmd := exec.Command("keybase", "fs", "ls", "-1", "--nocolor", path)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to list files in /keybase/team/: %s (%v)", strings.TrimSpace(string(output)), err)
	}
	var ret []string
	for _, s := range strings.Split(string(output), "\n") {
		if s != "" {
			ret = append(ret, s)
		}
	}
	return ret, nil
}
