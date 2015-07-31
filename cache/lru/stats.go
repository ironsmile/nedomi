/*
	This file contains the LRUCache's implementation of the CacheStats interface.
*/

package lru

import (
	"fmt"
	"github.com/ironsmile/nedomi/config"
	"github.com/ironsmile/nedomi/types"
)

/*
	LruCacheStats is used by the LRUCache to implement the CacheStats interface.
*/
type LruCacheStats struct {
	id       string
	hits     uint64
	requests uint64
	size     config.BytesSize
	objects  uint64
}

// CacheHitPrc implements part of CacheStats interface
func (lcs *LruCacheStats) CacheHitPrc() string {
	if lcs.requests == 0 {
		return ""
	}
	return fmt.Sprintf("%.f%%", (float32(lcs.Hits())/float32(lcs.Requests()))*100)
}

// ID implements part of CacheStats interface
func (lcs *LruCacheStats) ID() string {
	return lcs.id
}

// Hits implements part of CacheStats interface
func (lcs *LruCacheStats) Hits() uint64 {
	return lcs.hits
}

// Size implements part of CacheStats interface
func (lcs *LruCacheStats) Size() config.BytesSize {
	return lcs.size
}

// Objects implements part of CacheStats interface
func (lcs *LruCacheStats) Objects() uint64 {
	return lcs.objects
}

// Requests implements part of CacheStats interface
func (lcs *LruCacheStats) Requests() uint64 {
	return lcs.requests
}

// Stats implements part of CacheManager interface
func (l *LRUCache) Stats() types.CacheStats {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	var sum config.BytesSize
	var allObjects uint64

	for i := 0; i < cacheTiers; i++ {
		objects := config.BytesSize(l.tiers[i].Len())
		sum += (l.CacheZone.PartSize * objects)
		allObjects += uint64(objects)
	}

	return &LruCacheStats{
		id:       l.CacheZone.Path,
		hits:     l.hits,
		requests: l.requests,
		size:     sum,
		objects:  allObjects,
	}
}
