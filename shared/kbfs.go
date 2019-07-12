package shared

import (
	"fmt"
	"os/exec"
	"strings"
)

func KBFSFileExists(kbfsFilename string) (bool, error) {
	cmd := exec.Command("keybase", "fs", "stat", kbfsFilename)
	bytes, err := cmd.CombinedOutput()
	if err == nil {
		return true, nil
	}
	if strings.Contains(string(bytes), "ERROR file does not exist") {
		return false, nil
	}
	return false, fmt.Errorf("failed to stat %s: %s (%v)", kbfsFilename, string(bytes), err)
}

func KBFSRead(kbfsFilename string) ([]byte, error) {
	cmd := exec.Command("keybase", "fs", "read", kbfsFilename)
	bytes, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %s (%v)", kbfsFilename, string(bytes), err)
	}
	return bytes, nil
}

func KBFSDelete(filename string) error {
	cmd := exec.Command("keybase", "fs", "rm", filename)
	bytes, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to delete the file at %s: %s (%v)", filename, string(bytes), err)
	}
	return nil
}

func KBFSWrite(filename string, contents string, appendToFile bool) error {
	var cmd *exec.Cmd
	if appendToFile {
		cmd = exec.Command("keybase", "fs", "write", "--append")
	} else {
		cmd = exec.Command("keybase", "fs", "write")
	}

	cmd.Stdin = strings.NewReader(string(contents))
	bytes, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to write to file at %s: %s (%v)", filename, string(bytes), err)
	}
	return nil
}

func KBFSList(path string) ([]string, error) {
	cmd := exec.Command("keybase", "fs", "ls", "-1", "--nocolor", path)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to list files in /keybase/team/: %s (%v)", string(output), err)
	}
	var ret []string
	for _, s := range strings.Split(string(output), "\n") {
		if s != "" {
			ret = append(ret, s)
		}
	}
	return ret, nil
}
