package disk

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ironsmile/nedomi/config"
	"github.com/ironsmile/nedomi/logger"
	"github.com/ironsmile/nedomi/types"
	"github.com/ironsmile/nedomi/utils/testutils"
)

func getTestDiskStorage(t *testing.T, partSize int) (*Disk, string, func()) {
	diskPath, cleanup := testutils.GetTestFolder(t)

	d, err := New(&config.CacheZone{
		Path:     diskPath,
		PartSize: types.BytesSize(partSize),
	}, logger.NewMock())

	if err != nil {
		t.Fatalf("Could not create storage: %s", err)
	}

	return d, diskPath, cleanup
}

func TestRandomFilenameGeneration(t *testing.T) {
	t.Parallel()

	res1 := appendRandomSuffix("test")
	res2 := appendRandomSuffix("test")
	if res1 == res2 || !strings.Contains(res1, "test") || !strings.Contains(res2, "test") {
		t.Errorf("Invalid generated strings: %s, %s", res1, res2)
	}
}

func TestDiskPaths(t *testing.T) {
	t.Parallel()
	idx := &types.ObjectIndex{
		ObjID: types.NewObjectID("1.2", "/somewhere2"),
		Part:  33,
	}

	diskPath := "/some/path"
	disk := &Disk{path: diskPath}

	hash := idx.ObjID.StrHash()
	expectedHash := "052fb8b15a7737b1e7b70546b1c5023f0bd00a7d"
	if hash != expectedHash {
		t.Errorf("Incorrect ObjectID hash. Exected %s, got %s", expectedHash, hash)
	}

	objIDPath := disk.getObjectIDPath(idx.ObjID)
	expectedObjIDPath := filepath.Join(diskPath, idx.ObjID.CacheKey(), expectedHash[:2], expectedHash[2:4], expectedHash)
	if objIDPath != expectedObjIDPath {
		t.Errorf("Incorrect ObjectID path. Exected %s, got %s", expectedObjIDPath, objIDPath)
	}

	objIdxPath := disk.getObjectIndexPath(idx)
	expectedObjIdxPath := filepath.Join(expectedObjIDPath, "000033")
	if objIdxPath != expectedObjIdxPath {
		t.Errorf("Incorrect ObjectIndex path. Exected %s, got %s", expectedObjIdxPath, objIdxPath)
	}

	objMetadataPath := disk.getObjectMetadataPath(idx.ObjID)
	expectedObjMetadataPath := filepath.Join(expectedObjIDPath, objectMetadataFileName)
	if objMetadataPath != expectedObjMetadataPath {
		t.Errorf("Incorrect ObjectMetadata path. Exected %s, got %s", expectedObjMetadataPath, objMetadataPath)
	}
}

func TestFileCreation(t *testing.T) {
	t.Parallel()
	disk, diskPath, cleanup := getTestDiskStorage(t, 10)
	defer cleanup()

	filePath := filepath.Join(diskPath, "testdir1", "testdir2", "file")

	if _, err := disk.createFile(filePath); err != nil {
		t.Errorf("Error when creating the test file: %s", err)
	}

	if _, err := disk.createFile(filePath); !os.IsExist(err) {
		t.Errorf("Trying to create the same file again does not produce os.ErrExist: %v", err)
	}

	fileStat, err := os.Stat(filePath)
	if err != nil {
		t.Errorf("Cannot stat created file: %s", err)
	}

	if fileStat.Mode() != disk.filePermissions {
		t.Errorf("Desired and actual file permissions diverge: %s, %s", disk.filePermissions, fileStat.Mode())
	}

	dirStat, err := os.Stat(filepath.Dir(filePath))
	if err != nil {
		t.Errorf("Cannot stat created file's directory: %s", err)
	}

	if dirStat.Mode() != disk.dirPermissions {
		t.Errorf("Desired and actual directory permissions diverge: %s, %s", disk.dirPermissions, dirStat.Mode())
	}
}

func TestPartSizeCalculation(t *testing.T) {
	t.Parallel()
	type testcase struct {
		objSize        uint64
		partNum        uint32
		expectedResult uint64
	}

	for partSize := uint64(2); partSize <= 10; partSize++ {

		disk := &Disk{partSize: partSize}

		tests := []testcase{
			{partSize, 0, partSize},
			{partSize + 2, 1, 2},
			{partSize * 2, 0, partSize},
			{partSize * 2, 1, partSize},
			{partSize * 2, 2, 0},
			{partSize*2 + 1, 0, partSize},
			{partSize*2 + 1, 1, partSize},
			{partSize*2 + 1, 2, 1},
			{partSize*2 + 1, 3, 0},
		}

		for _, test := range tests {
			result := disk.getPartSize(test.partNum, test.objSize)
			if result != test.expectedResult {
				t.Errorf("Got part size %d with test %+v and partSize %d", result, test, partSize)
			}
		}
	}
}

func TestPartNumberValidation(t *testing.T) {
	t.Parallel()
	disk := &Disk{}

	for _, partNum := range []string{"asd", "-1", "12345", "12345a", "000111a"} {
		if _, err := disk.getPartNumberFromFile(partNum); err == nil {
			t.Errorf("Expected to receive error with invalid part number '%s'", partNum)
		}
	}

	for expNum, partNumStr := range []string{"000000", "000001", "000002"} {
		res, err := disk.getPartNumberFromFile(partNumStr)
		if err != nil {
			t.Errorf("Received error with valid part number '%s': %s", partNumStr, err)
		}
		if res != uint32(expNum) {
			t.Errorf("Expected part number %d and got %d", expNum, res)
		}
	}
}

func TestObjectMetadataLoading(t *testing.T) {
	t.Parallel()
	t.Skip("TODO: write tests for getObjectMetadata()")
}

func TestDiskSettingsLoadAndSave(t *testing.T) {
	t.Parallel()
	t.Skip("TODO: write tests for checkPreviousDiskSettings() and saveSettingsOnDisk()")
}
