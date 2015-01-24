package cache

import (
	"container/list"
	"fmt"
	"log"
	"sync"

	"github.com/gophergala/nedomi/config"
	. "github.com/gophergala/nedomi/types"
)

const (
	cacheTiers = 4
)

/*
   Implements LRU Cache
*/
type LRUCache struct {
	CacheZone *config.CacheZoneSection

	tiers  [cacheTiers]*list.List
	lookup map[ObjectIndex]*list.Element
	mutex  sync.Mutex

	tierListSize int
}

// Implements part of CacheManager interface
func (l *LRUCache) Has(oi ObjectIndex) bool {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	_, ok := l.lookup[oi]
	return ok
}

// Implements part of CacheManager interface
func (l *LRUCache) ObjectIndexStored(oi ObjectIndex) bool {
	err := l.AddObjectIndex(oi)
	if err != nil {
		log.Printf("Error storing object: %s", err)
		return false
	}
	return true
}

// Implements part of CacheManager interface
func (l *LRUCache) AddObjectIndex(oi ObjectIndex) error {
	if l.Has(oi) {
		return fmt.Errorf("Object already in cache: %s", oi)
	}

	l.mutex.Lock()
	defer l.mutex.Unlock()

	lastList := l.tiers[cacheTiers-1]

	l.lookup[oi] = lastList.PushFront(oi)

	if lastList.Len() > l.tierListSize {
		val := lastList.Remove(lastList.Back()).(ObjectIndex)
		l.remove(val)
		delete(l.lookup, val)
	}

	return nil
}

func (l *LRUCache) remove(oi ObjectIndex) {
	//!TODO: call the storage's Remove method
}

// Implements part of CacheManager interface
func (l *LRUCache) ConsumedSize() config.BytesSize {
	var sum config.BytesSize

	for i := 0; i < cacheTiers; i++ {
		sum += (l.CacheZone.PartSize * config.BytesSize(l.tiers[i].Len()))
	}

	return sum
}

// Implements part of CacheManager interface
func (l *LRUCache) Init() {
	for i := 0; i < cacheTiers; i++ {
		l.tiers[i] = list.New()
	}
	l.lookup = make(map[ObjectIndex]*list.Element)
	l.tierListSize = int(l.CacheZone.StorageObjects / uint64(cacheTiers))
}
