package config

import (
	"errors"

	"github.com/ironsmile/nedomi/types"
)

// CacheZone contains all configuration options for cache zones.
type CacheZone struct {
	ID                 string
	Type               string          `json:"type"`
	Path               string          `json:"path"`
	StorageObjects     uint64          `json:"storage_objects"`
	PartSize           types.BytesSize `json:"part_size"`
	Algorithm          string          `json:"cache_algorithm"`
	BulkRemoveCount    uint64          `json:"bulk_remove_count"`
	BulkRemoveTimeout  uint64          `json:"bulk_remove_timeout"`
	SkipCacheKeyInPath bool            `json:"skip_cache_key_in_path"`
}

// Validate checks a CacheZone config section for errors.
func (cz *CacheZone) Validate() error {
	//!TODO: support flexible type and config check for different modules
	if cz.ID == "" || cz.Type == "" || cz.Path == "" || cz.Algorithm == "" || cz.PartSize == 0 {
		return errors.New("missing or invalid information in the cache zone config section")
	}

	return nil
}

// GetSubsections returns nil (CacheZone has no subsections).
func (cz *CacheZone) GetSubsections() []Section {
	return nil
}
