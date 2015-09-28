// Package lru contains a LRU cache eviction implementation.
package lru

// !TODO: write about the tiered LRU

import (
	"container/list"
	"strconv"
	"sync"

	"github.com/ironsmile/nedomi/config"
	"github.com/ironsmile/nedomi/types"

	"github.com/timtadh/data-structures/trie"
	tstTypes "github.com/timtadh/data-structures/types"
)

const (
	// How many segments are there in the cache. 0 is the "best" segment in sense that
	// it contains the most recent files.
	cacheTiers int  = 4
	delimeter  byte = '#'
)

// Element is stored in the cache lookup hashmap
type Element struct {
	// Pointer to the linked list element
	ListElem *list.Element

	// In which tier this LRU element is. Tiers are from 0 up to cacheTiers
	ListTier int
}

//!TODO optimize
func cacheKeyAndPathToPath(key, path string) []byte {
	var result = make([]byte, 0, len(key)+len(path)+1) // not accurate
	result = append(result, []byte(key)...)
	result = append(result, delimeter)
	result = append(result, []byte(path)...)
	return result
}

func objectIDToPath(oid *types.ObjectID) []byte {
	var result = make([]byte, 0, len(oid.CacheKey())+len(oid.Path())+2) // not accurate
	result = append(result, []byte(oid.CacheKey())...)
	result = append(result, delimeter)
	result = append(result, []byte(oid.Path())...)
	result = append(result, delimeter)
	return result
}

func objectIndexToPath(index *types.ObjectIndex) []byte {
	var result = objectIDToPath(index.ObjID)
	return strconv.AppendUint(result, uint64(index.Part), 10)
}

// TieredLRUCache implements segmented LRU Cache. It has cacheTiers segments.
type TieredLRUCache struct {
	cfg *config.CacheZone

	tiers  [cacheTiers]*list.List
	lookup *trie.TST
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

	var ok = tc.lookup.Has(objectIndexToPath(oi))

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

	var oiPath = objectIndexToPath(oi)
	if ok := tc.lookup.Has(oiPath); ok {
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

	tc.logger.Logf("Storing %s in cache", oi)
	return tc.lookup.Put(oiPath, le)
}

// This function makes space for a new object in a full last list.
// In case there is space in the upper lists it puts its first element upwards.
// In case there is not - it removes its last element to make space.
func (tc *TieredLRUCache) freeSpaceInLastList() {
	lastListInd := cacheTiers - 1
	lastList := tc.tiers[lastListInd]

	if lastList.Len() < 1 {
		tc.logger.Log("Last list is empty but cache is trying to free space in it")
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
			valLruEl, err := tc.lookup.Get(objectIndexToPath(&val))
			if err != nil {
				tc.logger.Errorf("ERROR! Object in cache list was not found in the "+
					" lookup map: %v got err %s", val, err)
				i++
				continue
			}
			var elem = (valLruEl.(*Element))
			elem.ListElem = tc.tiers[i-1].PushBack(val)
		}
	} else {
		// There is no free slots anywhere in the upper tiers. So we will have to
		// remove something from the cache in order to make space.
		val := lastList.Remove(lastList.Back()).(types.ObjectIndex)
		tc.remove(&val)
		if _, err := tc.lookup.Remove(objectIndexToPath(&val)); err != nil {
			tc.logger.Errorf("ERROR! while removing %v from lookup trie - %s", val, err)

		}
	}
}

// Remove the object given from the cache returning true if it was in the cache
func (tc *TieredLRUCache) Remove(oi *types.ObjectIndex) bool {
	tc.mutex.Lock()
	defer tc.mutex.Unlock()

	var oiPath = objectIndexToPath(oi)
	if elI, err := tc.lookup.Get(oiPath); err == nil {
		tc.remove(oi) // !TODO check if needed
		if _, err = tc.lookup.Remove(oiPath); err != nil {
			tc.logger.Errorf(
				"got error while removing an element (for %v) that was just removed from the lookup trie - %s",
				oi, err)
		}
		el := elI.(*Element)
		tc.tiers[el.ListTier].Remove(el.ListElem)
		//!TODO reorder
		return true
	}
	return false
}

// RemoveObject the object given from the cache returning true if it was in the cache
func (tc *TieredLRUCache) RemoveObject(id *types.ObjectID) bool {
	tc.mutex.Lock()
	defer tc.mutex.Unlock()
	tc.logger.Logf("[%p] removing Object : %v from the cache", tc, id)
	defer tc.logger.Logf("[%p] finished removing Object : %v from the cache", tc, id)
	return tc.removeForPrefix(objectIDToPath(id))
}

// removeForPrefix removes all objects starting with the provided prefix. Returns false if nothing is deleted, true otherwise.
func (tc *TieredLRUCache) removeForPrefix(prefix []byte) bool {
	var oi types.ObjectIndex
	var key tstTypes.Hashable
	var val interface{}
	var el *Element

	var iterator = tc.lookup.PrefixFind(prefix)
	for key, val, iterator = iterator(); iterator != nil; key, val, iterator = iterator() {
		el = val.(*Element)
		oi = (el.ListElem.Value.(types.ObjectIndex))
		if _, err := tc.lookup.Remove(key.(tstTypes.ByteSlice)); err != nil {
			tc.logger.Errorf(
				"got error while removing an object index `%v` that was just retrived from the lookup trie - %s",
				oi, err)
		}
		tc.tiers[el.ListTier].Remove(el.ListElem)
	}

	return el != nil
}

func (tc *TieredLRUCache) remove(oi *types.ObjectIndex) {
	tc.logger.Logf("Removing %s from cache", oi)
	if err := tc.removeFunc(oi); err != nil {
		tc.logger.Errorf("Error removing %s from cache", oi)
	}
}

// PromoteObject implements part of types.CacheAlgorithm interface.
// It will reorder the linked lists so that this object index will be promoted in
// rank.
func (tc *TieredLRUCache) PromoteObject(oi *types.ObjectIndex) {
	tc.mutex.Lock()
	defer tc.mutex.Unlock()

	lruElI, err := tc.lookup.Get(objectIndexToPath(oi))

	if err != nil {
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

	var lruEl = lruElI.(*Element)
	if lruEl.ListTier == 0 {
		// This object is in the uppermost tier. It has nowhere to be promoted to
		// but the front of the tier.
		if tc.tiers[lruEl.ListTier].Front() == lruEl.ListElem {
			return
		}
		tc.tiers[lruEl.ListTier].MoveToFront(lruEl.ListElem)
		return
	}

	upperTier := tc.tiers[lruEl.ListTier-1]

	defer func() {
		lruEl.ListTier--
	}()

	if upperTier.Len() < tc.tierListSize {
		// The upper tier is not yet full. So we can push our object at the end
		// of it without needing to remove anything from it.
		tc.tiers[lruEl.ListTier].Remove(lruEl.ListElem)
		lruEl.ListElem = upperTier.PushFront(*oi)
		return
	}

	// The upper tier is full. An element from it will be swapped with the one
	// currently promted.
	upperListLastOi := upperTier.Remove(upperTier.Back()).(types.ObjectIndex)
	upperListLastLruElI, err := tc.lookup.Get(objectIndexToPath(&upperListLastOi))

	if err != nil {
		tc.logger.Errorf("ERROR! Cache incosistency. Element from the linked list "+
			"was not found in the lookup table (%s)",
			err)
	} else {
		var upperListLastLruEl = upperListLastLruElI.(*Element)
		upperListLastLruEl.ListElem = tc.tiers[lruEl.ListTier].PushFront(upperListLastOi)
	}

	tc.tiers[lruEl.ListTier].Remove(lruEl.ListElem)
	lruEl.ListElem = upperTier.PushFront(*oi)

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
	tc.lookup = new(trie.TST)
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
