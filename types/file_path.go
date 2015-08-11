package types

import "path/filepath"

// FilePath represents the string path to a file on the local filesystem. It is
// mainly used for flag arguments because it implements the flag.Value interface
// and resolves the absolute path to the file when the Set() method is used.
// Thus, if the current working directory is changed during the execution of the
// program, the path still remains valid.
type FilePath string

func (fp *FilePath) String() string {
	return string(*fp)
}

// Set converts the supplied argument to an absolute path. It does NOT check a
// file or folder exists at this path.
func (fp *FilePath) Set(path string) error {
	abs, err := filepath.Abs(path)
	*fp = FilePath(abs)
	return err
}
