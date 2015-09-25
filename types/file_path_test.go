package types

import (
	"os"
	"path/filepath"
	"testing"
)

func testCertainFile(t *testing.T, input, expectedOutput string) {
	var fp FilePath

	if err := fp.Set(input); err != nil {
		t.Errorf("Could not set file path '%s': %s", input, err)
	}
	if !filepath.IsAbs(string(fp)) {
		t.Errorf("The file '%s' path is not absolute", fp)
	}
	if string(fp) != expectedOutput {
		t.Errorf("Expected result '%s' but got '%s'", expectedOutput, fp)
	}
}

func TestFilePath(t *testing.T) {
	t.Parallel()
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Could not get the current working directory: %s", err)
	}

	testCertainFile(t, "simple.file", cwd+string(os.PathSeparator)+"simple.file")
	testCertainFile(t, "./simple.file", cwd+string(os.PathSeparator)+"simple.file")
	testCertainFile(t, "../relative.file", filepath.Dir(cwd)+string(os.PathSeparator)+"relative.file")
	testCertainFile(t, "./../relative.file", filepath.Dir(cwd)+string(os.PathSeparator)+"relative.file")
	testCertainFile(t, "/an/absolute/path", "/an/absolute/path")
	testCertainFile(t, "/a/somewhat/../relative/path", "/a/relative/path")
	testCertainFile(t, "///a/strange/absolute/path", "/a/strange/absolute/path")
}
