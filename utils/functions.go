// Package utils exports few handy functions
package utils

import (
	"os"

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
	//!TODO: implementation, tests
	return true
}
