package storage

import (
	"testing"

	"github.com/ironsmile/nedomi/config"
)

func TestCreatingBogusStorage(t *testing.T) {
	_, err := New(
		"bogus_storage",
		config.CacheZoneSection{},
		nil,
		nil,
	)

	if err == nil {
		t.Error("There was no error when creating bogus storage")
	}
}
