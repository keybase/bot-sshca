package shared

import (
	"os/user"
	"path/filepath"
	"strings"
)

// Returns the location of the public key associated with the given private key
func KeyPathToPubKey(keyPath string) string {
	return keyPath + ".pub"
}

// Returns the location of the signature associated with the given private key
func KeyPathToCert(keyPath string) string {
	return keyPath + "-cert.pub"
}

// Returns the location of the private key associated with the given public key
func PubKeyPathToKeyPath(pubKeyPath string) string {
	return strings.Replace(pubKeyPath, ".pub", "", 1)
}

// Expand out a path that starts with a tilde to be an absolute path
func ExpandPathWithTilde(path string) string {
	usr, _ := user.Current()
	if strings.HasPrefix(path, "~/") {
		path = filepath.Join(usr.HomeDir, path[2:])
	}
	return path
}
