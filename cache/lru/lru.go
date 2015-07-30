/*
	Package lru contains a LRU cache eviction implementation.

	//!TODO: write about the tiered LRU
*/
package lru

import (
	"container/list"
	"fmt"
	"log"
	"sync"

	"github.com/ironsmile/nedomi/config"
	. "github.com/ironsmile/nedomi/types"
)

const (
	// How many segments are there in the cache. 0 is the "best" segment in sense that
	// it contains the most recent files.
	cacheTiers int = 4
)

/*
   Struct which is stored in the cache lookup hashmap
*/
type LRUElement struct {
	// Pointer to the linked list element
	ListElem *list.Element

	// In which tier this LRU element is. Tiers are from 0 up to cacheTiers
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
func (l *LRUCache) Lookup(oi ObjectIndex) bool {
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
func (l *LRUCache) ShouldKeep(oi ObjectIndex) bool {
	err := l.AddObject(oi)
	if err != nil {
		log.Printf("Error storing object: %s", err)
		return true
	}
	return true
}

// Implements part of CacheManager interface
func (l *LRUCache) AddObject(oi ObjectIndex) error {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	if _, ok := l.lookup[oi]; ok {
		//!TODO: Create AlreadyInCacheErr type which implements the error interface
		return fmt.Errorf("Object already in cache: %s", oi)
	}

	lastList := l.tiers[cacheTiers-1]

	if lastList.Len() >= l.tierListSize {
		l.freeSpaceInLastList()
	}

	le := &LRUElement{
		ListTier: cacheTiers - 1,
		ListElem: lastList.PushFront(oi),
	}

	log.Printf("Storing %s in cache", oi)
	l.lookup[oi] = le

	return nil
}

/*
  This function makes space for a new object in a full last list.
  In case there is space in the upper lists it puts its first element upwards.
  In case there is not - it removes its last element to make space.
*/
func (l *LRUCache) freeSpaceInLastList() {
	lastListInd := cacheTiers - 1
	lastList := l.tiers[lastListInd]

	if lastList.Len() < 1 {
		log.Println("Last list is empty but cache is trying to free space in it")
		return
	}

	freeList := -1
	for i := lastListInd - 1; i >= 0; i-- {
		if l.tiers[i].Len() < l.tierListSize {
			freeList = i
			break
		}
	}

	if freeList != -1 {
		// There is a free space upwards in the list tiers. Move every front list
		// element to the back of the upper tier untill we reach this free slot.
		for i := lastListInd; i > freeList; i-- {
			front := l.tiers[i].Front()
			if front == nil {
				continue
			}
			val := l.tiers[i].Remove(front).(ObjectIndex)
			valLruEl, ok := l.lookup[val]
			if !ok {
				log.Printf("ERROR! Object in cache list was not found in the "+
					" lookup map: %v", val)
				i++
				continue
			}
			valLruEl.ListElem = l.tiers[i-1].PushBack(val)
		}
	} else {
		// There is no free slots anywhere in the upper tiers. So we will have to
		// remove something from the cache in order to make space.
		val := lastList.Remove(lastList.Back()).(ObjectIndex)
		l.remove(val)
		delete(l.lookup, val)
	}
}

func (l *LRUCache) remove(oi ObjectIndex) {
	log.Printf("Removing %s from cache", oi)
	if l.removeChan == nil {
		log.Println("Error! LRU cache is trying to write into empty remove channel.")
		return
	}
	l.removeChan <- oi
}

// Implements the CacheManager interface
func (l *LRUCache) ReplaceRemoveChannel(ch chan<- ObjectIndex) {
	l.removeChan = ch
}

/*
   Implements part of CacheManager interface.
   It will reorder the linked lists so that this object index will be promoted in
   rank.
*/
func (l *LRUCache) PromoteObject(oi ObjectIndex) {

	l.mutex.Lock()
	defer l.mutex.Unlock()

	lruEl, ok := l.lookup[oi]

	if !ok {
		// Unlocking the mutex in order to prevent a deadlock while calling
		// AddObject which tries to lock it too.
		l.mutex.Unlock()

		// This object is not in the cache yet. So we add it.
		if err := l.AddObject(oi); err != nil {
			log.Printf("Adding object in cache failed. Object: %v\n%s\n", oi, err)
		}

		// The mutex must be locked because of the deferred Unlock
		l.mutex.Lock()
		return
	}

	if lruEl.ListTier == 0 {
		// This object is in the uppermost tier. It has nowhere to be promoted to
		// but the front of the tier.
		if l.tiers[lruEl.ListTier].Front() == lruEl.ListElem {
			return
		}
		l.tiers[lruEl.ListTier].MoveToFront(lruEl.ListElem)
		return
	}

	upperTier := l.tiers[lruEl.ListTier-1]

	defer func() {
		lruEl.ListTier -= 1
	}()

	if upperTier.Len() < l.tierListSize {
		// The upper tier is not yet full. So we can push our object at the end
		// of it without needing to remove anything from it.
		l.tiers[lruEl.ListTier].Remove(lruEl.ListElem)
		lruEl.ListElem = upperTier.PushFront(oi)
		return
	}

	// The upper tier is full. An element from it will be swapped with the one
	// currently promted.
	upperListLastOi := upperTier.Remove(upperTier.Back()).(ObjectIndex)
	upperListLastLruEl, ok := l.lookup[upperListLastOi]

	if !ok {
		log.Println("ERROR! Cache incosistency. Element from the linked list " +
			"was not found in the lookup table")
	} else {
		upperListLastLruEl.ListElem = l.tiers[lruEl.ListTier].PushFront(upperListLastOi)
	}

	l.tiers[lruEl.ListTier].Remove(lruEl.ListElem)
	lruEl.ListElem = upperTier.PushFront(oi)

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

func (l *LRUCache) init() {
	for i := 0; i < cacheTiers; i++ {
		l.tiers[i] = list.New()
	}
	l.lookup = make(map[ObjectIndex]*LRUElement)
	l.tierListSize = int(l.CacheZone.StorageObjects / uint64(cacheTiers))
}

/*
	New returns LRUCache object ready for use.
*/
func New(cz *config.CacheZoneSection) *LRUCache {
	lru := &LRUCache{CacheZone: cz}
	lru.init()
	return lru
}
