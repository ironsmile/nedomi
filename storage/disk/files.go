package disk

import (
	"os"
	"path"
)

func CreateFile(filePath string) (*os.File, error) {
	if err := os.MkdirAll(path.Dir(filePath), 0700); err != nil {
		return nil, err
	}

	return os.Create(filePath)
}
