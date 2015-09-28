package lru

import (
	"fmt"
	"testing"

	"github.com/ironsmile/nedomi/config"
	"github.com/ironsmile/nedomi/logger"
	"github.com/ironsmile/nedomi/types"
	"github.com/timtadh/data-structures/trie"
)

func getCacheZone() *config.CacheZone {
	return &config.CacheZone{
		ID:             "default",
		Path:           "/some/path",
		StorageObjects: 30,
		PartSize:       2 * 1024 * 1024,
		Algorithm:      "lru",
	}
}

func getObjectIndex() *types.ObjectIndex {
	return &types.ObjectIndex{
		Part:  3,
		ObjID: types.NewObjectID("1.1", "/path"),
	}
}

func getObjectIndexFor(partNum uint32, cacheKey, path string) *types.ObjectIndex {
	return &types.ObjectIndex{
		Part:  partNum,
		ObjID: types.NewObjectID(cacheKey, path),
	}
}

func mockRemove(*types.ObjectIndex) error {
	return nil
}

func getFullLruCache(t *testing.T) *TieredLRUCache {
	cz := getCacheZone()
	lru := New(cz, mockRemove, logger.NewMock())

	storateObjects := (cz.StorageObjects / uint64(cacheTiers)) * uint64(cacheTiers)

	for i := uint64(0); i < storateObjects; i++ {

		oi := &types.ObjectIndex{
			Part:  uint32(i),
			ObjID: types.NewObjectID("1.1", "/path/to/many/objects"),
		}

		for k := 0; k < cacheTiers; k++ {
			lru.PromoteObject(oi)
		}
	}

	if objects := lru.Stats().Objects(); objects != storateObjects {
		t.Errorf("The cache was not full. Expected %d objects but it had %d",
			storateObjects, objects)
	}

	return lru
}

func TestLookupAndRemove(t *testing.T) {
	t.Parallel()
	cz := getCacheZone()
	oi := getObjectIndex()
	var removeCalled []*types.ObjectIndex
	lru := New(cz, func(oi *types.ObjectIndex) error {
		removeCalled = append(removeCalled, oi)
		return nil
	}, logger.NewMock())

	if lru.Lookup(oi) {
		t.Error("Empty LRU cache returned True for a object index lookup")
	}

	if err := lru.AddObject(oi); err != nil {
		t.Errorf("Error adding object into the cache. %s", err)
	}
	oi = getObjectIndex() // get a new/same objectIndex
	if !lru.Lookup(oi) {
		t.Error("Lookup for object index which was just added returned false")
	}

	if !lru.Remove(oi) {
		t.Error("Remove for object index which was just there returned false")
	}

	if lru.Lookup(oi) {
		t.Error("Lookup for object index which was just removed returned true")
	}

	if len(removeCalled) != 1 {
		t.Errorf("removeFunc was not called exactly once but %d times with the following arguments %+v", len(removeCalled), removeCalled)
	} else {
		if removeCalled[0] != oi {
			t.Errorf("removeFunc was not called with the expected argument %s but with %s", oi, removeCalled[0])
		}
	}

}

func TestSize(t *testing.T) {
	t.Parallel()
	cz := getCacheZone()
	oi := getObjectIndex()
	lru := New(cz, nil, logger.NewMock())

	if err := lru.AddObject(oi); err != nil {
		t.Errorf("Error adding object into the cache. %s", err)
	}

	if objects := lru.Stats().Objects(); objects != 1 {
		t.Errorf("Expec 1 object but found %d", objects)
	}

	if err := lru.AddObject(oi); err == nil {
		t.Error("Exepected error when adding object for the second time")
	}

	for i := 0; i < 16; i++ {
		oii := &types.ObjectIndex{
			Part:  uint32(i),
			ObjID: types.NewObjectID("1.1", "/path/to/other/object"),
		}

		if err := lru.AddObject(oii); err != nil {
			t.Errorf("Adding object in cache. %s", err)
		}
	}

	if objects := lru.Stats().Objects(); objects != 17 {
		t.Errorf("Expec 17 objects but found %d", objects)
	}

	if size, expected := lru.ConsumedSize(), 17*cz.PartSize; size != expected {
		t.Errorf("Expected total size to be %d but it was %d", expected, size)
	}
}

func getFromLookup(tst *trie.TST, index *types.ObjectIndex) (*Element, error) {
	elI, err := tst.Get(objectIndexToPath(index))
	if err != nil {
		return nil, err
	}
	return elI.(*Element), nil
}

func TestPromotionsInEmptyCache(t *testing.T) {
	t.Parallel()
	cz := getCacheZone()
	oi := getObjectIndex()
	lru := New(cz, nil, logger.NewMock())

	lru.PromoteObject(oi)

	if objects := lru.Stats().Objects(); objects != 1 {
		t.Errorf("Expected 1 object but found %d", objects)
	}

	lruEl, err := getFromLookup(lru.lookup, oi)

	if err != nil {
		t.Errorf("Was not able to find the object in the LRU table - %s", err)
	}

	if lruEl.ListTier != cacheTiers-1 {
		t.Errorf("Object was not in the last tier but in %d", lruEl.ListTier)
	}

	lru.PromoteObject(oi)

	if lruEl.ListTier != cacheTiers-2 {
		t.Errorf("Promoted object did not change its tier. "+
			"Expected it to be at tier %d but it was at %d", cacheTiers-2,
			lruEl.ListTier)
	}

	for i := 0; i < cacheTiers; i++ {
		lru.PromoteObject(oi)
	}

	if lruEl.ListTier != 0 {
		t.Errorf("Expected the promoted object to be in the uppermost "+
			"tier but it was at tier %d", lruEl.ListTier)
	}
}

func TestPromotionInFullCache(t *testing.T) {
	t.Parallel()

	lru := getFullLruCache(t)

	testOi := &types.ObjectIndex{
		Part:  0,
		ObjID: types.NewObjectID("1.1", "/path/to/tested/object"),
	}

	for currentTier := cacheTiers - 1; currentTier >= 0; currentTier-- {
		lru.PromoteObject(testOi)
		lruEl, err := getFromLookup(lru.lookup, testOi)
		if err != nil {
			t.Fatalf("Lost object while promoting it to tier %d - %s", currentTier, err)
		}

		if lruEl.ListTier != currentTier {
			t.Errorf("Tested LRU was not in the expected tier. It was suppsed to be"+
				" in tier %d but it was in %d", currentTier, lruEl.ListTier)
		}
	}

}

func TestShouldKeepMethod(t *testing.T) {
	t.Parallel()
	cz := getCacheZone()
	oi := getObjectIndex()
	lru := New(cz, nil, logger.NewMock())

	if shouldKeep := lru.ShouldKeep(oi); !shouldKeep {
		t.Error("LRU cache was supposed to return true for all ShouldKeep questions" +
			"but it returned false")
	}

	if objects := lru.Stats().Objects(); objects != 1 {
		t.Error("ShouldKeep was suppsed to add the object into the cache but id did not")
	}

	if shouldKeep := lru.ShouldKeep(oi); !shouldKeep {
		t.Error("ShouldKeep returned false after its second call")
	}

}

func TestPromotionToTheFrontOfTheList(t *testing.T) {
	t.Parallel()
	lru := getFullLruCache(t)

	testOiFirst := &types.ObjectIndex{
		Part:  0,
		ObjID: types.NewObjectID("1.1", "/path/to/tested/object"),
	}

	testOiSecond := &types.ObjectIndex{
		Part:  1,
		ObjID: types.NewObjectID("1.1", "/path/to/tested/object"),
	}

	for currentTier := cacheTiers - 1; currentTier >= 0; currentTier-- {
		lru.PromoteObject(testOiFirst)
		lru.PromoteObject(testOiSecond)
	}

	// First promoting the first object to the front of the list
	lru.PromoteObject(testOiFirst)

	lruEl, err := getFromLookup(lru.lookup, testOiFirst)

	if err != nil {
		t.Fatalf("Recently added object was not in the lookup table - %s", err)
	}

	if lru.tiers[0].Front() != lruEl.ListElem {
		t.Error("The expected element was not at the front of the top list")
	}

	// Then promoting the second one
	lru.PromoteObject(testOiSecond)

	lruEl, err = getFromLookup(lru.lookup, testOiSecond)

	if err != nil {
		t.Fatalf("Recently added object was not in the lookup table - %s", err)
	}

	if lru.tiers[0].Front() != lruEl.ListElem {
		t.Error("The expected element was not at the front of the top list")
	}
}

func TestRemoveForPath(t *testing.T) {
	cz := getCacheZone()
	var pathToFile = "/path/to/file"
	toBeRemoved := []*types.ObjectIndex{
		getObjectIndexFor(1, "2", pathToFile),
		getObjectIndexFor(3, "2", pathToFile),
		getObjectIndexFor(6, "2", pathToFile),
		getObjectIndexFor(2, "2", pathToFile+"pesho"),
	}

	notToBeRemoved := []*types.ObjectIndex{
		getObjectIndexFor(1, "1", pathToFile),
		getObjectIndexFor(2, "1", pathToFile),
		getObjectIndexFor(4, "1", pathToFile),
	}
	lru := New(cz, removeFunction(t, toBeRemoved), logger.NewMock())
	adAll(lru, toBeRemoved...)
	adAll(lru, notToBeRemoved...)

	lru.RemoveObjectsForKey("2", pathToFile)

	testRemoved(t, lru, toBeRemoved...)
	testNotRemoved(t, lru, notToBeRemoved...)
}

func TestRemoveObject(t *testing.T) {
	cz := getCacheZone()
	var pathToFile = "/path/to/file"
	toBeRemoved := []*types.ObjectIndex{
		getObjectIndexFor(1, "2", pathToFile),
		getObjectIndexFor(3, "2", pathToFile),
		getObjectIndexFor(6, "2", pathToFile),
	}

	notToBeRemoved := []*types.ObjectIndex{
		getObjectIndexFor(2, "2", pathToFile+"pesho"),
		getObjectIndexFor(1, "1", pathToFile),
		getObjectIndexFor(2, "1", pathToFile),
		getObjectIndexFor(4, "1", pathToFile),
	}

	lru := New(cz, removeFunction(t, toBeRemoved), logger.NewMock())
	adAll(lru, toBeRemoved...)
	adAll(lru, notToBeRemoved...)

	lru.RemoveObject(toBeRemoved[0].ObjID)

	testRemoved(t, lru, toBeRemoved...)
	testNotRemoved(t, lru, notToBeRemoved...)
}

func removeFunction(t *testing.T, toBeRemoved []*types.ObjectIndex) func(oi *types.ObjectIndex) error {
	return func(oi *types.ObjectIndex) error {
		for _, index := range toBeRemoved {
			if index.HashStr() == oi.HashStr() {
				return nil
			}
		}
		err := fmt.Errorf("objectIndex %v was removed but it shouldn't have been", oi)
		t.Error(err)
		return err
	}
}

func adAll(lru *TieredLRUCache, indexes ...*types.ObjectIndex) {
	for _, index := range indexes {
		lru.AddObject(index)
	}

}

func testRemoved(t *testing.T, lru *TieredLRUCache, toBeRemoved ...*types.ObjectIndex) {
	for _, index := range toBeRemoved {
		if ok := lru.Lookup(index); ok {
			t.Errorf("index `%v` was to be removed but it isn't", index)

		}
	}
}

func testNotRemoved(t *testing.T, lru *TieredLRUCache, notToBeRemoved ...*types.ObjectIndex) {
	for _, index := range notToBeRemoved {
		if ok := lru.Lookup(index); !ok {
			t.Errorf("index `%v` was to be not removed but it is", index)

		}
	}
}
