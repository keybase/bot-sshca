package shared

import (
	"io/ioutil"
	"os"
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

// Tries to read data from a file from given path.
// If no file is found, the file read creates an error or is empty,
// the function will return the value of the `or` parameter instead.
func ReadFileOrDefault(filePath string, or string) string {
	if len(filePath) <= 0 {
		return or
	}
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return or
	}
	data, err := ioutil.ReadFile(filePath)
	if err != nil || len(data) <= 0 {
		return or
	}
	return string(data)
}