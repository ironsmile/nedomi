package config

import (
	"encoding/json"
	"fmt"
	"net/url"
)

// BaseLocationSection contains the basic configuration options for virtual host's. location.
type BaseLocationSection struct {
	Match           string         `json:"match"`
	UpstreamType    string         `json:"upstream_type"`
	UpstreamAddress string         `json:"upstream_address"`
	CacheZone       string         `json:"cache_zone"`
	CacheKey        string         `json:"cache_key"`
	HandlerType     string         `json:"handler"`
	Logger          *LoggerSection `json:"logger"`
}

// LocationSection contains all configuration options for virtual host's location.
// It redefines some of the base fields to use the correct types.
type LocationSection struct {
	BaseLocationSection
	UpstreamAddress *url.URL          `json:"upstream_address"`
	CacheZone       *CacheZoneSection `json:"cache_zone"`
	parent          *VirtualHost
}

// UnmarshalJSON is a custom JSON unmashalling that also implements inheritance
// and custom field initiation
func (ls *LocationSection) UnmarshalJSON(buff []byte) error {
	// Parse the base values
	if err := json.Unmarshal(buff, &ls.BaseLocationSection); err != nil {
		return err
	}
	// Convert the upstream URL from string to url.URL
	if ls.BaseLocationSection.UpstreamAddress != "" {
		parsed, err := url.Parse(ls.BaseLocationSection.UpstreamAddress)
		if err != nil {
			return fmt.Errorf("Error parsing location %s upstream. %s", ls, err)
		}
		ls.UpstreamAddress = parsed
	}

	// Inject the cache zone configuration from the root config
	if cz, ok := ls.parent.parent.parent.CacheZones[ls.BaseLocationSection.CacheZone]; ok {
		ls.CacheZone = cz
	} else {
		return fmt.Errorf("Location %s has an invalid cache zone %s", ls, ls.CacheZone.ID)
	}

	return nil
}

// Validate checks the virtual host config for logical errors.
func (ls *LocationSection) Validate() error {
	if ls.Match == "" {
		return fmt.Errorf("All locations should have a match setting")
	}

	if ls.HandlerType == "" {
		return fmt.Errorf("Missing handler type for ls %s", ls)
	}

	return nil
}

func (ls *LocationSection) String() string {
	return fmt.Sprintf("%s.%s", ls.parent.Name, ls.Match)
}

// GetSubsections returns the ls config subsections.
func (ls *LocationSection) GetSubsections() []Section {
	return []Section{ls.Logger}
}
