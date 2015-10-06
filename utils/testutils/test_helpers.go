package testutils

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// GetTestFolder creates and returns a random test folder and a cleanup function.
// If the folder could not be created or removed afterwords, the test fails fatally.
func GetTestFolder(t testing.TB) (string, func()) {
	path, err := ioutil.TempDir("", "nedomi")
	if err != nil {
		t.Fatalf("Could not get a temporary folder: %s", err)
	}

	cleanup := func() {
		if err := os.RemoveAll(path); err != nil {
			t.Fatalf("Could delete the temp folder '%s': %s", path, err)
		}
	}

	return path, cleanup
}

// ProjectPath returns a path to the project source as an absolute directory name.
func ProjectPath() (string, error) {
	gopath := os.ExpandEnv("$GOPATH")
	relPath := filepath.FromSlash("src/github.com/ironsmile/nedomi")
	for _, path := range strings.Split(gopath, ":") {
		rootPath := filepath.Join(path, relPath)
		entry, err := os.Stat(rootPath)
		if err != nil {
			continue
		}

		if entry.IsDir() {
			return rootPath, nil
		}
	}

	return "", errors.New("Was not able to find the project path")
}
