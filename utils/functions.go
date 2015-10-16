// Package utils exports few handy functions
package utils

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ironsmile/nedomi/types"
)

// FileExists returns true if filePath is already existing regular file. If it is a
// directory FileExists will return false.
func FileExists(filePath string) bool {
	st, err := os.Stat(filePath)
	return err == nil && !st.IsDir()
}

// IsMetadataFresh checks whether the supplied metadata could still be used.
func IsMetadataFresh(obj *types.ObjectMetadata) bool {
	//!TODO: maybe make our own time package in which the time is cached. Calling
	// time.Now thousands of times per second does not look like a good idea.
	// This package can be made to work with a precision of one second and never
	// call time.Now more than that.
	return time.Unix(obj.ExpiresAt, 0).After(time.Now())
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
