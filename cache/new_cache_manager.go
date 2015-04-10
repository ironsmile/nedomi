/*
   This file contains the function which returns a new CacheManager object
   based on its string name.
*/

package cache

import (
	"fmt"

	"github.com/ironsmile/nedomi/cache/lru"
	"github.com/ironsmile/nedomi/config"
)

/*
   NewCacheManager creates and returns a particular type of cache manager.
*/
func NewCacheManager(ct string, cz *config.CacheZoneSection) (CacheManager, error) {

	cacheTypes := map[string]func(*config.CacheZoneSection) CacheManager{
		"lru": func(cz *config.CacheZoneSection) CacheManager {
			return lru.New(cz)
		},
	}

	fnc, ok := cacheTypes[ct]

	if !ok {
		return nil, fmt.Errorf("No such cache manager: `%s` type", ct)
	}

	return fnc(cz), nil
}
