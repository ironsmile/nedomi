// This file is generated with go generate. Any changes to it will be lost after
// subsequent generates.
// If you want to edit it go to types.go.template

package storage

import (
	"github.com/ironsmile/nedomi/cache"
	"github.com/ironsmile/nedomi/config"
	"github.com/ironsmile/nedomi/logger"
	"github.com/ironsmile/nedomi/upstream"

	"github.com/ironsmile/nedomi/storage/disk"
)

type newStorageFunc func(cfg config.CacheZoneSection, cm cache.Manager,
	up upstream.Upstream, log logger.Logger) Storage

var storageTypes = map[string]newStorageFunc{

	"disk": func(cfg config.CacheZoneSection, cm cache.Manager, up upstream.Upstream, log logger.Logger) Storage {
		return disk.New(cfg, cm, up, log)
	},
}
