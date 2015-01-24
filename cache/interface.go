package cache

import (
	"fmt"

	"github.com/gophergala/nedomi/config"
	"github.com/gophergala/nedomi/types"
)

/*
   CacheManager interface defines how a cache should behave
*/
type CacheManager interface {

	// Init is called only once after creating the CacheManager object
	Init()

	// Has returns wheather this object is in the cache or not
	Has(types.ObjectIndex) bool

	// ObjectIndexStored is called to signal that this ObjectIndex has been stored
	ObjectIndexStored(types.ObjectIndex) bool

	// AddObjectIndex adds this ObjectIndex to the cache
	AddObjectIndex(types.ObjectIndex) error

	// ConsumedSize returns the full size of all files currently in the cache
	ConsumedSize() config.BytesSize
}

/*
   NewCacheManager creates and returns a particular type of cache manager.
*/
func NewCacheManager(ct string, cz *config.CacheZoneSection) (CacheManager, error) {
	if ct != "lru" {
		return nil, fmt.Errorf("No such cache manager: `%s` type", ct)
	}
	return &LRUCache{CacheZone: cz}, nil
}
