// Package lru contains a LRU cache eviction implementation.
package lru

// !TODO: write about the tiered LRU

import (
	"container/list"
	"flag"
	"fmt"
	"runtime"
	"sync"
	"time"

	"github.com/ironsmile/nedomi/config"
	"github.com/ironsmile/nedomi/types"
)

var debug bool

func init() {
	flag.BoolVar(&debug, "check-lru", false, "do some additional checks in the lru cache algorithm(dev only)")
}

const (
	// How many segments are there in the cache. 0 is the "best" segment in sense that
	// it contains the most recent files.
	cacheTiers = 4
)

// Element is stored in the cache lookup hashmap
type Element struct {
	// Pointer to the linked list element
	ListElem *list.Element

	// In which tier this LRU element is. Tiers are from 0 up to cacheTiers
	ListTier int
}

// TieredLRUCache implements segmented LRU Cache. It has cacheTiers segments.
type TieredLRUCache struct {
	cfg *config.CacheZone

	tiers  [cacheTiers]*list.List
	lookup map[types.ObjectIndexHash]*Element
	mutex  sync.Mutex

	tierListSize int

	removeFunc func(*types.ObjectIndex) error

	logger types.Logger

	// Used to track cache hit/miss information
	requests uint64
	hits     uint64
}

// Lookup implements part of types.CacheAlgorithm interface
func (tc *TieredLRUCache) Lookup(oi *types.ObjectIndex) bool {
	tc.mutex.Lock()
	defer tc.mutex.Unlock()

	tc.requests++

	_, ok := tc.lookup[oi.Hash()]

	if ok {
		tc.hits++
	}

	return ok
}

// ShouldKeep implements part of types.CacheAlgorithm interface
func (tc *TieredLRUCache) ShouldKeep(oi *types.ObjectIndex) bool {
	if err := tc.AddObject(oi); err != nil && err != types.ErrAlreadyInCache {
		tc.logger.Errorf("Error storing object: %s", err)
	}
	return true
}

// AddObject implements part of types.CacheAlgorithm interface
func (tc *TieredLRUCache) AddObject(oi *types.ObjectIndex) error {
	tc.mutex.Lock()
	defer tc.mutex.Unlock()

	if _, ok := tc.lookup[oi.Hash()]; ok {
		return types.ErrAlreadyInCache
	}

	lastList := tc.tiers[cacheTiers-1]

	if lastList.Len() >= tc.tierListSize {
		tc.freeSpaceInLastList()
	}

	le := &Element{
		ListTier: cacheTiers - 1,
		ListElem: lastList.PushFront(*oi),
	}

	tc.logger.Debugf("Storing %s in lru", oi)
	tc.lookup[oi.Hash()] = le

	return nil
}

// This function makes space for a new object in a full last list.
// In case there is space in the upper lists it puts its first element upwards.
// In case there is not - it removes its last element to make space.
func (tc *TieredLRUCache) freeSpaceInLastList() {
	lastListInd := cacheTiers - 1
	lastList := tc.tiers[lastListInd]

	if lastList.Len() < 1 {
		tc.logger.Error("Last list is empty but cache is trying to free space in it")
		return
	}

	freeList := -1
	for i := lastListInd - 1; i >= 0; i-- {
		if tc.tiers[i].Len() < tc.tierListSize {
			freeList = i
			break
		}
	}

	if freeList != -1 {
		// There is a free space upwards in the list tiers. Move every front list
		// element to the back of the upper tier until we reach this free slot.
		for i := lastListInd; i > freeList; i-- {
			front := tc.tiers[i].Front()
			if front == nil {
				continue
			}
			val := tc.tiers[i].Remove(front).(types.ObjectIndex)
			valLruEl, ok := tc.lookup[val.Hash()]
			if !ok {
				tc.logger.Errorf("ERROR! Object in cache list was not found in the "+
					" lookup map: %v", val)
				i++
				continue
			}
			valLruEl.ListElem = tc.tiers[i-1].PushBack(val)
			valLruEl.ListTier = i - 1
		}
	} else {
		// There is no free slots anywhere in the upper tiers. So we will have to
		// remove something from the cache in order to make space.
		val := lastList.Remove(lastList.Back()).(types.ObjectIndex)
		delete(tc.lookup, val.Hash())
		if err := tc.removeFunc(&val); err != nil {
			tc.logger.Logf("error while removing %s from cache - %s", &val, err)
		}
	}
}

// Remove the objects given from the cache.
func (tc *TieredLRUCache) Remove(ois ...*types.ObjectIndex) {
	tc.mutex.Lock()
	defer tc.mutex.Unlock()

	for _, oi := range ois {
		if el, ok := tc.lookup[oi.Hash()]; ok {
			delete(tc.lookup, oi.Hash())
			tc.tiers[el.ListTier].Remove(el.ListElem)
		}
	}
}

// PromoteObject implements part of types.CacheAlgorithm interface.
// It will reorder the linked lists so that this object index will be promoted in
// rank.
func (tc *TieredLRUCache) PromoteObject(oi *types.ObjectIndex) {
	tc.mutex.Lock()
	defer tc.mutex.Unlock()
	if debug {
		tc.checkTiers()
		defer tc.checkTiers()
	}

	lruEl, ok := tc.lookup[oi.Hash()]

	if !ok {
		// Unlocking the mutex in order to prevent a deadlock while calling
		// AddObject which tries to lock it too.
		tc.mutex.Unlock()

		// This object is not in the cache yet. So we add it.
		if err := tc.AddObject(oi); err != nil {
			tc.logger.Errorf("Adding object in cache failed. Object: %v\n%s", oi, err)
		}

		// The mutex must be locked because of the deferred Unlock
		tc.mutex.Lock()
		return
	}

	currentTier := tc.tiers[lruEl.ListTier]
	if lruEl.ListTier == 0 {
		// This object is in the uppermost tier. It has nowhere to be promoted to
		// but the front of the tier.
		if currentTier.Front() == lruEl.ListElem {
			return
		}
		currentTier.MoveToFront(lruEl.ListElem)
		return
	}

	upperTier := tc.tiers[lruEl.ListTier-1]

	defer func() {
		currentTier.Remove(lruEl.ListElem)
		lruEl.ListElem = upperTier.PushFront(*oi)
		lruEl.ListTier--
	}()

	if upperTier.Len() < tc.tierListSize {
		// The upper tier is not yet full. So we can push our object at the end
		// of it without needing to remove anything from it.
		return
	}

	// The upper tier is full. An element from it will be swapped with the one
	// currently promoted.
	upperListLastOi := upperTier.Remove(upperTier.Back()).(types.ObjectIndex)
	upperListLastLruEl, ok := tc.lookup[upperListLastOi.Hash()]

	if !ok {
		tc.logger.Error("ERROR! Cache inconsistency. Element from the linked list " +
			"was not found in the lookup table")
		return
	}

	upperListLastLruEl.ListElem = currentTier.PushFront(upperListLastOi)
	upperListLastLruEl.ListTier = lruEl.ListTier
}

func (tc *TieredLRUCache) checkTiers() {
	for i := 0; i < cacheTiers; i++ {
		if tc.tiers[i].Len() > tc.tierListSize {
			tc.logger.Error(i, tc.tiers[i].Len())
			panic("tiers are not accurately sized")
		}
	}
}

// ConsumedSize implements part of types.CacheAlgorithm interface
func (tc *TieredLRUCache) ConsumedSize() types.BytesSize {
	tc.mutex.Lock()
	defer tc.mutex.Unlock()

	return tc.consumedSize()
}

func (tc *TieredLRUCache) consumedSize() types.BytesSize {
	var sum types.BytesSize

	for i := 0; i < cacheTiers; i++ {
		sum += (tc.cfg.PartSize * types.BytesSize(tc.tiers[i].Len()))
	}

	return sum
}

func (tc *TieredLRUCache) init() {
	for i := 0; i < cacheTiers; i++ {
		tc.tiers[i] = list.New()
	}
	tc.lookup = make(map[types.ObjectIndexHash]*Element)
	tc.tierListSize = int(tc.cfg.StorageObjects / uint64(cacheTiers))
}

// New returns TieredLRUCache object ready for use.
func New(cz *config.CacheZone, removeFunc func(*types.ObjectIndex) error,
	logger types.Logger) *TieredLRUCache {

	lru := &TieredLRUCache{
		cfg:        cz,
		removeFunc: removeFunc,
		logger:     logger,
	}
	lru.init()
	return lru
}

// ChangeConfig changes the TieredLRUCache config and start using it
func (tc *TieredLRUCache) ChangeConfig(bulkRemoveTimout, bulkRemoveCount, newsize uint64, logger types.Logger) {
	tc.mutex.Lock()
	defer tc.mutex.Unlock()
	tc.logger = logger
	tc.cfg.StorageObjects = newsize
	tc.cfg.BulkRemoveCount = bulkRemoveCount
	tc.cfg.BulkRemoveTimeout = bulkRemoveTimout
	tc.resize()
}

// resize the lru
func (tc *TieredLRUCache) resize() {
	if debug {
		defer tc.checkTiers()
	}

	var newtierListSize = int(tc.cfg.StorageObjects / 4)
	if tc.tierListSize > newtierListSize {
		var oids = tc.resizeDown(int(tc.stats().Objects() - tc.cfg.StorageObjects))

		var ch = make(chan struct{})
		go func() {
			defer close(ch)
			for _, oi := range oids {
				delete(tc.lookup, oi.Hash())
			}
		}()

		// for each tier from the upper most without the last
		for i := 0; cacheTiers-1 > i; i++ {
			// while it's bigger than the new size
			for tc.tiers[i].Len() > newtierListSize {
				// move it's last element to the lower tier
				tc.tiers[i+1].PushFront(tc.tiers[i].Remove(tc.tiers[i].Back()))
			}
		}

		var (
			additionalOids []types.ObjectIndex
			last           = tc.tiers[cacheTiers-1]
		)
		for last.Len() > newtierListSize {
			additionalOids = append(additionalOids, last.Remove(last.Back()).(types.ObjectIndex))
		}
		<-ch

		for _, oi := range additionalOids {
			delete(tc.lookup, oi.Hash())
		}

		go tc.throttledRemove(append(oids, additionalOids...))
	}
	tc.tierListSize = newtierListSize
}

// remove the elements with time inbetween removes,
// but only if they are not in the cache at the time of removal
func (tc *TieredLRUCache) throttledRemove(indexes []types.ObjectIndex) {
	defer func() {
		if msg := recover(); msg != nil {
			const size = 64 << 10
			buf := make([]byte, size)
			buf = buf[:runtime.Stack(buf, false)]
			tc.logger.Errorf(
				"Panic during throttled remove after resize down: %v\n%s",
				msg, buf)
			if debug {
				panic(fmt.Sprintf("%v\n%s", msg, buf))
			}
		}
	}()
	var timer = time.NewTimer(0)
	for i, n := 0, len(indexes); n > i; i += int(tc.cfg.BulkRemoveCount) {
		tc.removeIfMissing(indexes[i:min(i+int(tc.cfg.BulkRemoveCount), n)]...)
		timer.Reset(time.Duration(tc.cfg.BulkRemoveTimeout) * time.Millisecond)
		<-timer.C
	}
}

func min(l, r int) int {
	if l > r {
		return r
	}
	return l
}

func (tc *TieredLRUCache) removeIfMissing(ois ...types.ObjectIndex) {
	tc.mutex.Lock()
	defer tc.mutex.Unlock()

	for _, oi := range ois {
		if _, ok := tc.lookup[oi.Hash()]; !ok {
			tc.removeFunc(&oi)
		}
	}
}

func (tc *TieredLRUCache) resizeDown(remove int) []types.ObjectIndex {
	if 0 >= remove {
		return nil
	}
	var result = make([]types.ObjectIndex, remove)
	lastListInd := cacheTiers - 1
	var removed = 0
	var i int

	for i = lastListInd; i >= 0 && remove != removed; i-- {
		removed += removeFromList(tc.tiers[i], remove-removed, result[removed:])
	}

	return result
}

// removes up to n elements from the list starting backwards and putting their
// values in the removed slice (which should be atleast remove big). Also returns how
// many were removed
func removeFromList(l *list.List, remove int, removed []types.ObjectIndex) int {
	var e = l.Back()
	var prev *list.Element
	var i = 0
	for ; remove > i && e != nil; i++ {
		prev = e.Prev()
		removed[i] = l.Remove(e).(types.ObjectIndex)
		e = prev
	}

	return i
}
