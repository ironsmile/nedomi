package config

import (
	"encoding/json"
	"fmt"
	"net/url"
	"time"
)

// baseLocation contains the basic configuration options for virtual host's. location.
type baseLocation struct {
	Name                  string
	UpstreamType          string    `json:"upstream_type"`
	UpstreamAddress       string    `json:"upstream_address"`
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
	UpstreamAddress      *url.URL      `json:"upstream_address"`
	CacheZone            *CacheZone    `json:"cache_zone"`
	CacheDefaultDuration time.Duration `json:"cache_default_duration"`
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

	// Convert the upstream URL from string to url.URL
	if ls.baseLocation.UpstreamAddress != "" {
		parsed, err := url.Parse(ls.baseLocation.UpstreamAddress)
		if err != nil {
			return fmt.Errorf("Error parsing location %s upstream. %s", ls, err)
		}
		ls.UpstreamAddress = parsed
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
