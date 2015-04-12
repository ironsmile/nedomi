/*
   This file is generated with go generate. Any changes to it will be lost after
   subsequent generates.

   If you want to edit it go to cache_types.go.template
*/

package cache

import (
	"github.com/ironsmile/nedomi/config"

	"github.com/ironsmile/nedomi/cache/lru"
)

var cacheTypes map[string]func(*config.CacheZoneSection) CacheManager = map[string]func(
	*config.CacheZoneSection) CacheManager{

	"lru": func(cz *config.CacheZoneSection) CacheManager {
		return lru.New(cz)
	},
}
