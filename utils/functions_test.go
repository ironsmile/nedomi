package utils

import (
	"io/ioutil"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/ironsmile/nedomi/types"
	"github.com/ironsmile/nedomi/utils/testutils"
)

func TestFileExistsFunction(t *testing.T) {
	t.Parallel()
	tmpDir, cleanup := testutils.GetTestFolder(t)
	defer cleanup()

	if exists := FileExists(tmpDir); exists {
		t.Errorf("Expected false when calling FileExists with directory: %s", tmpDir)
	}

	tmpFile, err := ioutil.TempFile(tmpDir, "functest")

	if err != nil {
		t.Fatalf("Creating a temporary file filed: %s", err)
	}

	defer func() {
		tmpFile.Close()
		os.Remove(tmpFile.Name())
	}()

	if exists := FileExists(tmpFile.Name()); !exists {
		t.Errorf("Expected true when calling FileEists with a file %s", tmpFile.Name())
	}
}

var freshMetadataTests = []struct {
	expires time.Duration
	isFresh bool
}{
	{
		expires: time.Hour,
		isFresh: true,
	},
	{
		expires: -time.Hour,
		isFresh: false,
	},
	{
		expires: -time.Second,
		isFresh: false,
	},
	{
		expires: time.Second,
		isFresh: true,
	},
	{
		expires: -2 * time.Second,
		isFresh: false,
	},
	{
		expires: 2 * time.Second,
		isFresh: true,
	},
}

func TestIsMetadataFresh(t *testing.T) {

	for index, test := range freshMetadataTests {

		now := time.Now()

		obj := &types.ObjectMetadata{
			ResponseTimestamp: now.Unix(),
			Code:              200,
			Size:              535,
			Headers:           make(http.Header),
			ExpiresAt:         now.Add(test.expires).Unix(),
		}

		found := IsMetadataFresh(obj)

		if found != test.isFresh {
			t.Errorf("Test %d: expected %t for duration %s but got %t",
				index, test.isFresh, test.expires, found)
		}

	}
}
