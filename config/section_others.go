package config

import "encoding/json"

//TODO: split in different files, implement validation?

// CacheZoneSection contains all configuration options for cache zones.
type CacheZoneSection struct {
	ID             uint32    `json:"id"`
	Path           string    `json:"path"`
	StorageObjects uint64    `json:"storage_objects"`
	PartSize       BytesSize `json:"part_size"`
	CacheAlgo      string    `json:"cache_algorithm"`
}

// LoggerSection contains logger options
type LoggerSection struct {
	Type     string          `json:"type"`
	Settings json.RawMessage `json:"settings"`
}

// SystemSection contains system and environment configurations.
type SystemSection struct {
	Pidfile string `json:"pidfile"`
	Workdir string `json:"workdir"`
	User    string `json:"user"`
}
