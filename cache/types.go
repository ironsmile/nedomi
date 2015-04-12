/*
   This file is generated with go generate. Any changes to it will be lost after
   subsequent generates.

   If you want to edit it go to types.go.template
*/

package cache

import (
	"github.com/ironsmile/nedomi/config"

	"github.com/ironsmile/nedomi/cache/lru"
)

type newCacheFunc func(*config.CacheZoneSection) CacheManager

var cacheTypes map[string]newCacheFunc = map[string]newCacheFunc{

	"lru": func(cz *config.CacheZoneSection) CacheManager {
		return lru.New(cz)
	},
}
