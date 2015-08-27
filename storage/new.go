// Package storage deals with files on the disk or whatever storage. It defines the
// Storage interface. It has methods for getting contents of a file, headers of a
// file and methods for removing files. Since every cache zone has its own storage
// it is possible to have different storage implementations running at the same
// time.

// The storageTypes map is in types.go and it is generate with `go generate`.

//go:generate go run ../tools/module_generator/main.go -template "types.go.template" -output "types.go"

package storage

import (
	"fmt"

	"github.com/ironsmile/nedomi/config"
	"github.com/ironsmile/nedomi/types"
)

// New returns a new Storage ready for use. The st argument sets which type of
// storage will be returned.
func New(cfg *config.CacheZoneSection, log types.Logger) (types.Storage, error) {

	if cfg == nil {
		return nil, fmt.Errorf("Empty cache zone configuration supplied!")
	}

	storFunc, ok := storageTypes[cfg.Type]

	if !ok {
		return nil, fmt.Errorf("No such storage type: %s", cfg.Type)
	}

	return storFunc(cfg, log), nil
}
