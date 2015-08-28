package disk

import (
	"path"
	"testing"

	"github.com/ironsmile/nedomi/config"
	"github.com/ironsmile/nedomi/types"
)

func TestDiskPaths(t *testing.T) {
	idx := &types.ObjectIndex{
		ObjID: &types.ObjectID{
			CacheKey: "1.2",
			Path:     "/somewhere",
		},
		Part: 33,
	}

	diskPath := "/some/path"
	disk := New(&config.CacheZoneSection{Path: diskPath}, nil)

	hash := idx.ObjID.StrHash()
	expectedHash := "583fae38a17840864d328e08b0d21cec293f74b2"
	if hash != expectedHash {
		t.Errorf("Incorrect ObjectID hash. Exected %s, got %s", expectedHash, hash)
	}

	objIDPath := disk.getObjectIDPath(idx.ObjID)
	expectedObjIDPath := path.Join(diskPath, expectedHash[:2], expectedHash[2:4], expectedHash)
	if objIDPath != expectedObjIDPath {
		t.Errorf("Incorrect ObjectID path. Exected %s, got %s", expectedObjIDPath, objIDPath)
	}

	objIdxPath := disk.getObjectIndexPath(idx)
	expectedObjIdxPath := path.Join(expectedObjIDPath, "000033")
	if objIdxPath != expectedObjIdxPath {
		t.Errorf("Incorrect ObjectIndex path. Exected %s, got %s", expectedObjIdxPath, objIdxPath)
	}
}
