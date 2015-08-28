package disk

import (
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/ironsmile/nedomi/config"
	"github.com/ironsmile/nedomi/types"
)

// This creates and returns a random test folder and a cleanup function. If the
// folder could not be created or removed afterwords, the test fails fatally.
func getTestFolder(t *testing.T) (string, func()) {
	path, err := ioutil.TempDir("", "nedomi")
	if err != nil {
		t.Fatalf("Could not get a temporary folder: %s", err)
	}

	cleanup := func() {
		if err := os.RemoveAll(path); err != nil {
			t.Fatalf("Could delete the temp folder '%s': %s", path, err)
		}
	}

	return path, cleanup
}

func TestDiskPaths(t *testing.T) {
	t.Parallel()
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

	objMetadata := &types.ObjectMetadata{ID: idx.ObjID}
	objMetadataPath := disk.getObjectMetadataPath(objMetadata)
	expectedObjMetadataPath := path.Join(expectedObjIDPath, objectMetadataFileName)
	if objMetadataPath != expectedObjMetadataPath {
		t.Errorf("Incorrect ObjectMetadata path. Exected %s, got %s", expectedObjMetadataPath, objMetadataPath)
	}
}

func TestFileCreation(t *testing.T) {
	t.Parallel()
	diskPath, cleanup := getTestFolder(t)
	defer cleanup()

	filePath := path.Join(diskPath, "testdir1", "testdir2", "file")
	disk := New(&config.CacheZoneSection{Path: diskPath}, nil)

	if _, err := disk.createFile(filePath); err != nil {
		t.Errorf("Error when creating the test file: %s", err)
	}

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Errorf("Created file does not exist: %s", err)
	}

	//!TODO: test for errors when creating the same file twice?
	//!TODO: write test that check file and folder permissions
}
