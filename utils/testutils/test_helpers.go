package testutils

import (
	"io/ioutil"
	"os"
	"testing"
)

// GetTestFolder creates and returns a random test folder and a cleanup function.
// If the folder could not be created or removed afterwords, the test fails fatally.
func GetTestFolder(t *testing.T) (string, func()) {
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
