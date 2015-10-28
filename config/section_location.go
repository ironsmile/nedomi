package config

import (
	"encoding/json"
	"fmt"
	"time"
)

// baseLocation contains the basic configuration options for virtual host's. location.
type baseLocation struct {
	HeadersRewrite
	Name                  string
	Upstream              string    `json:"upstream"`
	CacheZone             string    `json:"cache_zone"`
	CacheKey              string    `json:"cache_key"`
	CacheDefaultDuration  string    `json:"cache_default_duration"`
	Handlers              []Handler `json:"handlers"`
	Logger                Logger    `json:"logger"`
	CacheKeyIncludesQuery bool      `json:"cache_key_includes_query"`
}

// Location contains all configuration options for virtual host's location.
type Location struct {
	baseLocation
	CacheZone            *CacheZone
	CacheDefaultDuration time.Duration
	parent               *VirtualHost
}

// UnmarshalJSON is a custom JSON unmashalling that also implements inheritance
// and custom field initiation
func (ls *Location) UnmarshalJSON(buff []byte) error {
	// Parse the baseLocation values
	if err := json.Unmarshal(buff, &ls.baseLocation); err != nil {
		return err
	}

	// Convert the location string to time.Duration
	if ls.baseLocation.CacheDefaultDuration == "" {
		if ls.parent != nil {
			// Inject the CacheDefaultDuration from the location's parent
			ls.CacheDefaultDuration = ls.parent.CacheDefaultDuration
		} else {
			//!TODO: maybe add HTTPSection-wide configuration option for default caching duration
			// and use it here instead of this hardcoded time.
			ls.CacheDefaultDuration = DefaultCacheDuration
		}
	} else if dur, err := time.ParseDuration(ls.baseLocation.CacheDefaultDuration); err != nil {
		return fmt.Errorf("Error parsing %s's cache_default_location: %s", ls, err)
	} else {
		ls.CacheDefaultDuration = dur
	}

	// Inject the cache zone configuration from the root config
	if cz, ok := ls.parent.parent.parent.CacheZones[ls.baseLocation.CacheZone]; ok {
		ls.CacheZone = cz
	} else {
		return fmt.Errorf("Location %s has an invalid cache zone `%s`", ls,
			ls.baseLocation.CacheZone)
	}

	return nil
}

// Validate checks the virtual host location config for logical errors.
func (ls *Location) Validate() error {
	if ls.Name == "" {
		return fmt.Errorf("All locations should have a match setting")
	}

	if len(ls.Handlers) == 0 {
		return fmt.Errorf("Missing handlers for location %s", ls)
	}

	if ls.CacheDefaultDuration <= 0 {
		return fmt.Errorf("Cache default duration in %s must be positive", ls)
	}

	return nil
}

func (ls *Location) String() string {
	if ls.parent == nil {
		return ls.Name
	}

	return fmt.Sprintf("%s.%s", ls.parent.Name, ls.Name)
}

// GetSubsections returns the ls config subsections.
func (ls *Location) GetSubsections() []Section {
	res := []Section{ls.Logger}
	for _, handler := range ls.Handlers {
		res = append(res, handler)
	}

	return res
}
