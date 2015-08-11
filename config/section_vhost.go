package config

import (
	"encoding/json"
	"fmt"
	"net/url"
)

// BaseVirtualHost contains the basic configuration options for virtual hosts.
type BaseVirtualHost struct {
	Name            string         `json:"name"`
	UpstreamType    string         `json:"upstream_type"`
	UpstreamAddress string         `json:"upstream_address"`
	CacheZone       string         `json:"cache_zone"`
	CacheKey        string         `json:"cache_key"`
	HandlerType     string         `json:"handler"`
	Logger          *LoggerSection `json:"logger"`
}

// VirtualHost contains all configuration options for virtual hosts. It
// redefines some of the base fields to use the correct types.
type VirtualHost struct {
	BaseVirtualHost
	UpstreamAddress *url.URL          `json:"upstream_address"`
	CacheZone       *CacheZoneSection `json:"cache_zone"`
	parent          *HTTP
}

// UnmarshalJSON is a custom JSON unmashalling that also implements inheritance
// and custom field initiation
func (vh *VirtualHost) UnmarshalJSON(buff []byte) error {
	// Parse the base values
	if err := json.Unmarshal(buff, &vh.BaseVirtualHost); err != nil {
		return err
	}
	// Convert the upstream URL from string to url.URL
	if vh.BaseVirtualHost.UpstreamAddress != "" {
		parsed, err := url.Parse(vh.BaseVirtualHost.UpstreamAddress)
		if err != nil {
			return fmt.Errorf("Error parsing server %s upstream. %s", vh.Name, err)
		}
		vh.UpstreamAddress = parsed
	}

	// Inject the cache zone configuration from the root config
	if cz, ok := vh.parent.parent.CacheZones[vh.BaseVirtualHost.CacheZone]; ok {
		vh.CacheZone = cz
	} else {
		return fmt.Errorf("Vhost %s has an invalid cache zone %s", vh.Name, vh.CacheZone.ID)
	}

	return nil
}

// Validate checks the virtual host config for logical errors.
func (vh *VirtualHost) Validate() error {
	if vh.Name == "" {
		return fmt.Errorf("All virtual hosts should have a name setting")
	}

	if vh.HandlerType == "" {
		return fmt.Errorf("Missing handler type for vhost %s", vh.Name)
	}

	//!TODO: support flexible type and config check for different modules
	if vh.HandlerType == "proxy" {
		if vh.UpstreamType == "" || vh.CacheKey == "" || vh.UpstreamAddress == nil {
			return fmt.Errorf("Missing required settings for vhost %s", vh.Name)
		}

		if !vh.UpstreamAddress.IsAbs() {
			return fmt.Errorf("Upstream address for server %s was not absolute: %s",
				vh.Name, vh.UpstreamAddress)
		}
	}

	return nil
}

// GetSubsections returns the vhost config subsections.
func (vh *VirtualHost) GetSubsections() []Section {
	return []Section{vh.Logger}
}
