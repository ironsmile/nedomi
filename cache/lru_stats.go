package cache

import (
	"github.com/gophergala/nedomi/config"
)

// Implements part of CacheManager interface
func (l *LRUCache) Stats() *CacheStats {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	var sum config.BytesSize
	var allObjects uint64

	for i := 0; i < cacheTiers; i++ {
		objects := config.BytesSize(l.tiers[i].Len())
		sum += (l.CacheZone.PartSize * objects)
		allObjects += uint64(objects)
	}

	return &CacheStats{
		ID:       l.CacheZone.Path,
		Hits:     l.hits,
		Requests: l.requests,
		Size:     sum,
		Objects:  allObjects,
	}
}
