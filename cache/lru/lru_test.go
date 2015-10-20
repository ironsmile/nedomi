package lru

import (
	"testing"

	"github.com/ironsmile/nedomi/config"
	"github.com/ironsmile/nedomi/mock"
	"github.com/ironsmile/nedomi/types"
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

func mockRemove(*types.ObjectIndex) error {
	return nil
}

func getFullLruCache(t *testing.T) *TieredLRUCache {
	cz := getCacheZone()
	lru := New(cz, mockRemove, mock.NewLogger())

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
	lru := New(cz, nil, mock.NewLogger())

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

	lru.Remove(oi)

	if lru.Lookup(oi) {
		t.Error("Lookup for object index which was just removed returned true")
	}
}

func TestSize(t *testing.T) {
	t.Parallel()
	cz := getCacheZone()
	oi := getObjectIndex()
	lru := New(cz, nil, mock.NewLogger())

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

func TestPromotionsInEmptyCache(t *testing.T) {
	t.Parallel()
	cz := getCacheZone()
	oi := getObjectIndex()
	lru := New(cz, nil, mock.NewLogger())

	lru.PromoteObject(oi)

	if objects := lru.Stats().Objects(); objects != 1 {
		t.Errorf("Expected 1 object but found %d", objects)
	}

	lruEl, ok := lru.lookup[oi.Hash()]

	if !ok {
		t.Error("Was not able to find the object in the LRU table")
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
		lruEl, ok := lru.lookup[testOi.Hash()]
		if !ok {
			t.Fatalf("Lost object while promoting it to tier %d", currentTier)
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
	lru := New(cz, nil, mock.NewLogger())

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

	lruEl, ok := lru.lookup[testOiFirst.Hash()]

	if !ok {
		t.Fatal("Recently added object was not in the lookup table")
	}

	if lru.tiers[0].Front() != lruEl.ListElem {
		t.Error("The expected element was not at the front of the top list")
	}

	// Then promoting the second one
	lru.PromoteObject(testOiSecond)

	lruEl, ok = lru.lookup[testOiSecond.Hash()]

	if !ok {
		t.Fatal("Recently added object was not in the lookup table")
	}

	if lru.tiers[0].Front() != lruEl.ListElem {
		t.Error("The expected element was not at the front of the top list")
	}
}

func TestResize(t *testing.T) {
	t.Parallel()

	lru := getFullLruCache(t)

	testOi := &types.ObjectIndex{
		Part:  0,
		ObjID: types.NewObjectID("1.1", "/path/to/tested/object"),
	}
	oldSize := lru.Stats().Objects()
	lru.Resize(oldSize + 20)
	lru.PromoteObject(testOi)
	if lru.Stats().Objects() != oldSize+1 {
		t.Errorf("It was expected that after resize more objects could be added but that wasn't true")
	}
}
