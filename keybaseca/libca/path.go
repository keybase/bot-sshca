package libca

import (
	"os/user"
	"path/filepath"
	"strings"
)

func ExpandPathWithTilde(path string) string {
	usr, _ := user.Current()
	if strings.HasPrefix(path, "~/") {
		path = filepath.Join(usr.HomeDir, path[2:])
	}
	return path
}
