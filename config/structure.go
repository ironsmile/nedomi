package config

import (
	"encoding/json"
	"net/url"
)

// Config is the root configuration type. It contains representation for
// everything in config.json.
type Config struct {
	System     SystemSection       `json:"system"`
	Logger     LoggerSection       `json:"logger"`
	HTTP       HTTPSection         `json:"http"`
	CacheZones []*CacheZoneSection `json:"cache_zones"`
}

// HTTPSection contains all configuration options for HTTP.
type HTTPSection struct {
	Listen         string         `json:"listen"`
	Servers        []*VirtualHost `json:"virtual_hosts"`
	MaxHeadersSize int            `json:"max_headers_size"`
	ReadTimeout    uint32         `json:"read_timeout"`
	WriteTimeout   uint32         `json:"write_timeout"`
	CacheAlgo      string         `json:"cache_algorithm"`
	UpstreamType   string         `json:"upstream_type"`
}

// VirtualHost contains all configuration options for virtual hosts.
type VirtualHost struct {
	Name            string         `json:"name"`
	UpstreamAddress string         `json:"upstream_address"`
	CacheZone       uint32         `json:"cache_zone"`
	CacheKey        string         `json:"cache_key"`
	HandlerType     string         `json:"handler"`
	UpstreamType    string         `json:"upstream_type"`
	Logger          *LoggerSection `json:"logger"`

	// used internally
	upstreamAddressURL *url.URL
	cacheZone          *CacheZoneSection
}

// CacheZoneSection contains all configuration options for cache zones.
type CacheZoneSection struct {
	ID             uint32    `json:"id"`
	Path           string    `json:"path"`
	StorageObjects uint64    `json:"storage_objects"`
	PartSize       BytesSize `json:"part_size"`
	CacheAlgo      string    `json:"cache_algorithm"`
}

// UpstreamURL returns the previously calculated *url.URL of the upstream
// attached to this VirtualHost.
func (vh *VirtualHost) UpstreamURL() *url.URL {
	return vh.upstreamAddressURL
}

// GetCacheZoneSection returns config.CacheZoneSection for this virtual host.
func (vh *VirtualHost) GetCacheZoneSection() *CacheZoneSection {
	return vh.cacheZone
}

// IsForProxyModule returns true if the virtual host should use the default
// proxy handler module as its handler. False otherwise.
func (vh *VirtualHost) IsForProxyModule() bool {
	return vh.HandlerType == "" || vh.HandlerType == "proxy"
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
