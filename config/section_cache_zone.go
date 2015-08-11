package config

import (
	"errors"

	"github.com/ironsmile/nedomi/types"
)

// CacheZoneSection contains all configuration options for cache zones.
type CacheZoneSection struct {
	ID             string
	Type           string          `json:"type"`
	Path           string          `json:"path"`
	StorageObjects uint64          `json:"storage_objects"`
	PartSize       types.BytesSize `json:"part_size"`
	Algorithm      string          `json:"cache_algorithm"`
}

// Validate checks a CacheZone config section for errors.
func (cz *CacheZoneSection) Validate() error {
	//!TODO: support flexible type and config check for different modules
	if cz.ID == "" || cz.Type == "" || cz.Path == "" || cz.Algorithm == "" || cz.PartSize == 0 {
		return errors.New("Missing or invalid information in the cache zone config section.")
	}

	return nil
}

// GetSubsections returns nil (CacheZoneSection has no subsections).
func (cz *CacheZoneSection) GetSubsections() []Section {
	return nil
}
