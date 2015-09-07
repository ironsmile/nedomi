package utils

import (
	"io/ioutil"
	"os"
	"testing"
)

func TestFileExistsFunction(t *testing.T) {
	tmpDir := os.TempDir()

	defer os.Remove(tmpDir)

	if exists := FileExists(tmpDir); exists {
		t.Errorf("Expected false when calling FileExists with directory: %s", tmpDir)
	}

	tmpFile, err := ioutil.TempFile(tmpDir, "functest")

	if err != nil {
		t.Fatalf("Creating a temporary file filed: %s", err)
	}

	defer func() {
		tmpFile.Close()
		os.Remove(tmpFile.Name())
	}()

	if exists := FileExists(tmpFile.Name()); !exists {
		t.Errorf("Expected true when calling FileEists with a file %s", tmpFile.Name())
	}
}
