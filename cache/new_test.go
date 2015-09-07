package cache

import (
	"os"
	"testing"

	"github.com/ironsmile/nedomi/config"
	"github.com/ironsmile/nedomi/logger"
	"github.com/ironsmile/nedomi/types"
)

func TestCreatingCacheAlgorithms(t *testing.T) {
	cz := config.CacheZoneSection{
		ID:             "default",
		Path:           os.TempDir(),
		PartSize:       4123123,
		StorageObjects: 9813743,
		Algorithm:      "lru",
	}

	if _, err := New(&cz, make(chan *types.ObjectIndex), logger.NewMock()); err != nil {
		t.Errorf("Error when creating cache algorithm. %s", err)
	}
}

func TestCreatingBogusCacheAlgorithmReturnsError(t *testing.T) {
	cz := config.CacheZoneSection{
		ID:             "default",
		Path:           os.TempDir(),
		PartSize:       4123123,
		StorageObjects: 9813743,
		Algorithm:      "bogus",
	}

	if _, err := New(&cz, make(chan *types.ObjectIndex), logger.NewMock()); err == nil {
		t.Error("Expected an error when creating bogus algorithm but got none")
	}
}
