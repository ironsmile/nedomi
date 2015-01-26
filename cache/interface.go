/*
	Package cache is implements the caching algorithm. It defines the CacheManager
	interface. Every CacheZone has its own cache manager. This makes it possible for
	different caching algorithms to be used in the same time.
*/
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

	// UsedObjectIndex is called every time this part of a file has been used
	// to satisfy a client request
	UsedObjectIndex(types.ObjectIndex)

	// ConsumedSize returns the full size of all files currently in the cache
	ConsumedSize() config.BytesSize

	// ReplaceRemoveChannel makes this cache communicate its desire to remove objects
	// on this channel
	ReplaceRemoveChannel(chan<- types.ObjectIndex)

	// Stats returns statistics for this cache manager
	Stats() *CacheStats
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
