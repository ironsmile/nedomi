package testutils

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
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
			time.Sleep(time.Second)                      // wait a while
			if err2 := os.RemoveAll(path); err2 != nil { // try again
				t.Fatalf("Could delete the temp folder '%s' twice: \n%s\n%s", path, err, err2)
			}
		}
	}

	return path, cleanup
}

// ShouldntFail checks if any of the supplied parameters are non-nil errors and
// it fatally fails the test if they are.
func ShouldntFail(t testing.TB, errors ...error) {
	for idx, err := range errors {
		if err != nil {
			t.Fatalf("An unexpected error occured in statement %d: %s", idx+1, err)
		}
	}
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
