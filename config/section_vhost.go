package config

import (
	"encoding/json"
	"fmt"
	"net/url"
)

// VirtualHostBase contains the basic configuration options for virtual hosts.
type VirtualHostBase struct {
	Name            string         `json:"name"`
	UpstreamType    string         `json:"upstream_type"`
	UpstreamAddress string         `json:"upstream_address"`
	CacheAlgo       string         `json:"cache_algorithm"`
	CacheZone       uint32         `json:"cache_zone"`
	CacheKey        string         `json:"cache_key"`
	HandlerType     string         `json:"handler"`
	Logger          *LoggerSection `json:"logger"`
}

// VirtualHost contains all configuration options for virtual hosts. It
// redefines some of the base fields to use the correct types.
type VirtualHost struct {
	VirtualHostBase
	UpstreamAddress *url.URL          `json:"upstream_address"`
	CacheZone       *CacheZoneSection `json:"cache_zone"`
	parent          *HTTPSection
}

// UnmarshalJSON is a custom JSON unmashalling that also implements inheritance
// and custom field initiation
func (vh *VirtualHost) UnmarshalJSON(buff []byte) error {
	// Parse the base values
	if err := json.Unmarshal(buff, &vh.VirtualHostBase); err != nil {
		return err
	}

	//!TODO: maybe leave the validation and other stuff to the handler, like how
	// we do it with the loggers...
	// Or modules may have 2 helper functions: New() and ParseConfig()?
	// Or better yet: unify the config parsing, validation and initialization
	// of the different objects.That way we only have a single pass
	// that either fails or succeeds before starting the application itself.

	// Convert the upstream URL from string to url.URL
	if vh.VirtualHostBase.UpstreamAddress != "" {
		parsed, err := url.Parse(vh.VirtualHostBase.UpstreamAddress)
		if err != nil {
			return fmt.Errorf("Error parsing server %s upstream. %s", vh.Name, err)
		}
		vh.UpstreamAddress = parsed
	}

	// Inject the cache zone configuration from the root config
	for _, cz := range vh.parent.parent.CacheZones {
		if cz.ID == vh.VirtualHostBase.CacheZone {
			vh.CacheZone = cz
		}
	}
	//!TODO: use somthing like this instead of the above code
	// (after making the cache zones into a map at the root level)
	/*
		if cz, ok := cacheZonesMap[vh.CacheZone]; ok {
			vh.cacheZone = cz
		} else {
			return fmt.Errorf("Upstream %s has not existing cache zone id. %d",
				vh.Name, vh.CacheZone)
		}
	*/

	return vh.Validate()
}

// Validate checks the virtual host config for logical errors.
func (vh *VirtualHost) Validate() error {
	//!TODO: implement

	if vh.UpstreamAddress != nil {
		if !vh.UpstreamAddress.IsAbs() {
			return fmt.Errorf("Upstream address for server %s was not absolute: %s",
				vh.Name, vh.VirtualHostBase.UpstreamAddress)
		}
	}

	return nil
}

// UpstreamURL returns the previously calculated *url.URL of the upstream
// attached to this VirtualHost.
//!TODO: remove
func (vh *VirtualHost) UpstreamURL() *url.URL {
	return vh.UpstreamAddress
}

// GetCacheZoneSection returns config.CacheZoneSection for this virtual host.
//!TODO: remove
func (vh *VirtualHost) GetCacheZoneSection() *CacheZoneSection {
	return vh.CacheZone
}

// IsForProxyModule returns true if the virtual host should use the default
// proxy handler module as its handler. False otherwise.
//!TODO: remove
func (vh *VirtualHost) IsForProxyModule() bool {
	return vh.HandlerType == "" || vh.HandlerType == "proxy"
}
