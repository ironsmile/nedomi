// This file is generated with go generate. Any changes to it will be lost after
// subsequent generates.
// If you want to edit it go to types.go.template

package storage

import (
	"github.com/ironsmile/nedomi/config"
	"github.com/ironsmile/nedomi/types"

	"github.com/ironsmile/nedomi/storage/disk"
)

type newStorageFunc func(cfg *config.CacheZoneSection, log types.Logger) (types.Storage, error)

var storageTypes = map[string]newStorageFunc{

	"disk": func(cfg *config.CacheZoneSection, log types.Logger) (types.Storage, error) {
		return disk.New(cfg, log)
	},
}
