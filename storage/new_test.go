package storage

import (
	"testing"

	"github.com/ironsmile/nedomi/config"
)

func TestCreatingBogusStorage(t *testing.T) {
	_, err := New(
		&config.CacheZone{Type: "bogus_storage"},
		nil,
	)

	if err == nil {
		t.Error("There was no error when creating bogus storage")
	}
}
