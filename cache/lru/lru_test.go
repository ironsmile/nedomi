package lru

import (
	"testing"

	"github.com/ironsmile/nedomi/config"
	"github.com/ironsmile/nedomi/types"
)

func getCacheZone() *config.CacheZoneSection {
	return &config.CacheZoneSection{
		ID:             1,
		Path:           "/some/path",
		StorageObjects: 30,
		PartSize:       2 * 1024 * 1024,
		CacheAlgo:      "lru",
	}
}

func getObjectIndex() types.ObjectIndex {
	return types.ObjectIndex{
		Part: 3,
		ObjID: types.ObjectID{
			CacheKey: "1.1",
			Path:     "/path",
		},
	}
}

func TestLRULookup(t *testing.T) {
	cz := getCacheZone()

	oi := getObjectIndex()

	lru := New(cz)

	if lru.Lookup(oi) {
		t.Error("Empty LRU cache returned True for a object index lookup")
	}

	if err := lru.AddObject(oi); err != nil {
		t.Errorf("Error adding object into the cache. %s", err)
	}

	if !lru.Lookup(oi) {
		t.Error("Lookup for object index which was just added returned false")
	}
}
