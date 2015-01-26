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
	cacheTiers int = 4
)

type LRUElement struct {
	ListElem *list.Element
	ListTier int
}

/*
   Implements segmented LRU Cache. It has cacheTiers segments.
*/
type LRUCache struct {
	CacheZone *config.CacheZoneSection

	tiers  [cacheTiers]*list.List
	lookup map[ObjectIndex]*LRUElement
	mutex  sync.Mutex

	tierListSize int

	removeChan chan<- ObjectIndex

	// Used to track cache hit/miss information
	requests uint64
	hits     uint64
}

// Implements part of CacheManager interface
func (l *LRUCache) Has(oi ObjectIndex) bool {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	l.requests += 1

	_, ok := l.lookup[oi]

	if ok {
		l.hits += 1
	}

	return ok
}

// Implements part of CacheManager interface
func (l *LRUCache) ObjectIndexStored(oi ObjectIndex) bool {
	err := l.AddObjectIndex(oi)
	if err != nil {
		log.Printf("Error storing object: %s", err)
		return true
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

	le := &LRUElement{
		ListTier: cacheTiers - 1,
		ListElem: lastList.PushFront(oi),
	}

	log.Printf("Storing %s in cache", oi)
	l.lookup[oi] = le

	if lastList.Len() > l.tierListSize {
		val := lastList.Remove(lastList.Back()).(ObjectIndex)
		l.remove(val)
		delete(l.lookup, val)
	}

	return nil
}

func (l *LRUCache) remove(oi ObjectIndex) {
	log.Printf("Removinb %s from cache", oi)
	if l.removeChan == nil {
		log.Println("Error! LRU cache is trying to write into empty remove channel.")
		return
	}
	l.removeChan <- oi
}

func (l *LRUCache) ReplaceRemoveChannel(ch chan<- ObjectIndex) {
	l.removeChan = ch
}

/*
   Implements part of CacheManager interface.
   It will reorder the linke lists so that this object index will be get promoted in
   rank.
*/
func (l *LRUCache) UsedObjectIndex(oi ObjectIndex) {

	l.mutex.Lock()
	defer l.mutex.Unlock()

	lruEl, ok := l.lookup[oi]

	if !ok {
		return
	}

	//!TODO: .Front will be equal to this elem extremely rarely. We will have to
	// devise a method to move to the upper tier if it was in the top N of this.
	if l.tiers[lruEl.ListTier].Front() == lruEl.ListElem {
		if lruEl.ListTier == 0 {
			return
		}

		upperTier := l.tiers[lruEl.ListTier-1]

		if upperTier.Len() < 1 {
			lruEl.ListElem = upperTier.PushBack(oi)
			return
		}

		upperListLastOi := upperTier.Remove(upperTier.Back()).(ObjectIndex)
		upperListLastLruEl, ok := l.lookup[upperListLastOi]

		if !ok {
			log.Println("ERROR! Cache incosistency. Element from the linked list " +
				"was not found in the lookup table")
			return
		}

		l.tiers[lruEl.ListTier].Remove(lruEl.ListElem)
		upperListLastLruEl.ListElem = l.tiers[lruEl.ListTier].PushFront(upperListLastOi)

		lruEl.ListElem = upperTier.PushBack(oi)
	} else {
		l.tiers[lruEl.ListTier].MoveToFront(lruEl.ListElem)
	}

}

// Implements part of CacheManager interface
func (l *LRUCache) ConsumedSize() config.BytesSize {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	return l.consumedSize()
}

func (l *LRUCache) consumedSize() config.BytesSize {
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
	l.lookup = make(map[ObjectIndex]*LRUElement)
	l.tierListSize = int(l.CacheZone.StorageObjects / uint64(cacheTiers))

}
