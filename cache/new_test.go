package cache

import (
	"testing"

	"github.com/ironsmile/nedomi/config"
	"github.com/ironsmile/nedomi/logger"
	"github.com/ironsmile/nedomi/types"
)

func mockRemove(*types.ObjectIndex) error {
	return nil
}

func TestCreatingCacheAlgorithms(t *testing.T) {
	t.Parallel()
	cz := config.CacheZone{
		ID:             "default",
		Path:           "/does/not/matter",
		PartSize:       4123123,
		StorageObjects: 9813743,
		Algorithm:      "lru",
	}

	if _, err := New(&cz, mockRemove, logger.NewMock()); err != nil {
		t.Errorf("Error when creating cache algorithm. %s", err)
	}
}

func TestCreatingBogusCacheAlgorithmReturnsError(t *testing.T) {
	t.Parallel()
	cz := config.CacheZone{
		ID:             "default",
		Path:           "/does/not/matter",
		PartSize:       4123123,
		StorageObjects: 9813743,
		Algorithm:      "bogus",
	}

	if _, err := New(&cz, mockRemove, logger.NewMock()); err == nil {
		t.Error("Expected an error when creating bogus algorithm but got none")
	}
}

func TestCreatingCacheAlgorithmWithNilConfigReturnsError(t *testing.T) {
	t.Parallel()
	if _, err := New(nil, mockRemove, logger.NewMock()); err == nil {
		t.Error("Expected an error when creating bogus algorithm but got none")
	}
}
