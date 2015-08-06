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
	UpstreamAddress url.URL           `json:"upstream_address"`
	CacheZone       *CacheZoneSection `json:"cache_zone"`
	parent          *HTTPSection
}

// UnmarshalJSON is a custom JSON unmashalling that also implements inheritance
// and custom field initiation
func (b *VirtualHost) UnmarshalJSON(buff []byte) error {
	// Parse the base values
	if err := json.Unmarshal(buff, &b.VirtualHostBase); err != nil {
		return err
	}

	//!TODO: maybe leave the validation and other stuff to the handler, like how
	// we do it with the loggers...
	// Or better yet: unify the config parsing, validation and initialization
	// of the different objects.That way we only have a single pass
	// that either fails or succeeds before starting the application itself.

	if b.VirtualHostBase.UpstreamAddress != "" {
		// Convert the upstream URL from string to url.URL
		parsed, err := url.Parse(b.VirtualHostBase.UpstreamAddress)
		if err != nil {
			return fmt.Errorf("Error parsing server %s upstream. %s", b.Name, err)
		}
		if !parsed.IsAbs() {
			return fmt.Errorf("Upstream address for server %s was not absolute: %s",
				b.Name, b.VirtualHostBase.UpstreamAddress)
		}
		b.UpstreamAddress = *parsed
	}

	return b.Validate()
}

func (b *VirtualHost) Validate() error {
	//!TODO: implement
	return nil
}

//!TODO: remove
// UpstreamURL returns the previously calculated *url.URL of the upstream
// attached to this VirtualHost.
func (vh *VirtualHost) UpstreamURL() *url.URL {
	return &vh.UpstreamAddress
}

//!TODO: remove
// GetCacheZoneSection returns config.CacheZoneSection for this virtual host.
func (vh *VirtualHost) GetCacheZoneSection() *CacheZoneSection {
	return vh.CacheZone
}

//!TODO: remove
// IsForProxyModule returns true if the virtual host should use the default
// proxy handler module as its handler. False otherwise.
func (vh *VirtualHost) IsForProxyModule() bool {
	return vh.HandlerType == "" || vh.HandlerType == "proxy"
}
