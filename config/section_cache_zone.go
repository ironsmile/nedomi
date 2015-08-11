package config

import "github.com/ironsmile/nedomi/types"

// CacheZoneSection contains all configuration options for cache zones.
type CacheZoneSection struct {
	ID             string
	Type           string          `json:"type"`
	Path           string          `json:"path"`
	StorageObjects uint64          `json:"storage_objects"`
	PartSize       types.BytesSize `json:"part_size"`
	Algorithm      string          `json:"cache_algorithm"`
}
