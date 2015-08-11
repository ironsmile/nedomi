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

	//!TODO: maybe leave the validation and other stuff to the handler, like how
	// we do it with the loggers...
	// Or modules may have 2 helper functions: New() and ParseConfig()?
	// Or better yet: unify the config parsing, validation and initialization
	// of the different objects.That way we only have a single pass
	// that either fails or succeeds before starting the application itself.

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

	return vh.Validate()
}

// Validate checks the virtual host config for logical errors.
func (vh *VirtualHost) Validate() error {
	//!TODO: implement

	if vh.UpstreamAddress != nil {
		if !vh.UpstreamAddress.IsAbs() {
			return fmt.Errorf("Upstream address for server %s was not absolute: %s",
				vh.Name, vh.BaseVirtualHost.UpstreamAddress)
		}
	}

	return nil
}