/*
   This file contains the function which returns a new CacheManager object
   based on its string name.

   NewCacheManager uses the cacheTypes map. This map is generated with
   `go generate` in the cache_types.go file.
*/

//go:generate ./generate_cache_types

package cache

import (
	"fmt"

	"github.com/ironsmile/nedomi/config"
)

/*
   NewCacheManager creates and returns a particular type of cache manager.
*/
func NewCacheManager(ct string, cz *config.CacheZoneSection) (CacheManager, error) {

	fnc, ok := cacheTypes[ct]

	if !ok {
		return nil, fmt.Errorf("No such cache manager: `%s` type", ct)
	}

	return fnc(cz), nil
}

/*
   Returns true if a CacheManager with this name exists. False otherwise.
*/
func CacheManagerTypeExists(ct string) bool {
	_, ok := cacheTypes[ct]
	return ok
}
