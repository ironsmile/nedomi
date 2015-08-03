package lru

// This file contains the LRUCache's implementation of the CacheStats interface.

import (
	"fmt"

	"github.com/ironsmile/nedomi/config"
	"github.com/ironsmile/nedomi/types"
)

// TieredCacheStats is used by the LRUCache to implement the CacheStats interface.
type TieredCacheStats struct {
	id       string
	hits     uint64
	requests uint64
	size     config.BytesSize
	objects  uint64
}

// CacheHitPrc implements part of CacheStats interface
func (lcs *TieredCacheStats) CacheHitPrc() string {
	if lcs.requests == 0 {
		return ""
	}
	return fmt.Sprintf("%.f%%", (float32(lcs.Hits())/float32(lcs.Requests()))*100)
}

// ID implements part of CacheStats interface
func (lcs *TieredCacheStats) ID() string {
	return lcs.id
}

// Hits implements part of CacheStats interface
func (lcs *TieredCacheStats) Hits() uint64 {
	return lcs.hits
}

// Size implements part of CacheStats interface
func (lcs *TieredCacheStats) Size() config.BytesSize {
	return lcs.size
}

// Objects implements part of CacheStats interface
func (lcs *TieredCacheStats) Objects() uint64 {
	return lcs.objects
}

// Requests implements part of CacheStats interface
func (lcs *TieredCacheStats) Requests() uint64 {
	return lcs.requests
}

// Stats implements part of cache.Manager interface
func (tc *TieredLRUCache) Stats() types.CacheStats {
	tc.mutex.Lock()
	defer tc.mutex.Unlock()

	var sum config.BytesSize
	var allObjects uint64

	for i := 0; i < cacheTiers; i++ {
		objects := config.BytesSize(tc.tiers[i].Len())
		sum += (tc.CacheZone.PartSize * objects)
		allObjects += uint64(objects)
	}

	return &TieredCacheStats{
		id:       tc.CacheZone.Path,
		hits:     tc.hits,
		requests: tc.requests,
		size:     sum,
		objects:  allObjects,
	}
}
