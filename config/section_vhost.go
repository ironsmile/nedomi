package config

import (
	"encoding/json"
	"fmt"
	"net/url"
)

// baseVirtualHost contains the basic configuration options for virtual hosts.
type baseVirtualHost struct {
	Locations map[string]json.RawMessage `json:"locations"`
}

// VirtualHost contains all configuration options for virtual hosts. It
// redefines some of the baseLocation fields to use the correct types.
type VirtualHost struct {
	baseVirtualHost
	Location
	Locations []*Location `json:"locations"`
	parent    *HTTP
}

// UnmarshalJSON is a custom JSON unmashalling that also implements inheritance
// and custom field initiation
func (vh *VirtualHost) UnmarshalJSON(buff []byte) error {
	// Parse the baseLocation values
	if err := json.Unmarshal(buff, &vh.baseLocation); err != nil {
		return err
	}
	if err := json.Unmarshal(buff, &vh.baseVirtualHost); err != nil {
		return err
	}

	// Convert the upstream URL from string to url.URL
	if vh.baseLocation.UpstreamAddress != "" {
		parsed, err := url.Parse(vh.baseLocation.UpstreamAddress)
		if err != nil {
			return fmt.Errorf("Error parsing server %s upstream. %s", vh.Name, err)
		}
		vh.UpstreamAddress = parsed
	}

	// Inject the cache zone configuration from the root config
	vh.CacheZone = vh.parent.parent.CacheZones[vh.baseLocation.CacheZone]

	locationBase := Location{
		parent: vh,
		baseLocation: baseLocation{
			Handlers:        append([]Handler(nil), vh.Handlers...),
			UpstreamType:    vh.UpstreamType,
			UpstreamAddress: vh.baseLocation.UpstreamAddress,
			CacheZone:       vh.baseLocation.CacheZone,
			CacheKey:        vh.baseLocation.CacheKey,
			Logger:          vh.Logger,
		},
	}

	// Parse all the locations
	for match, locationBuff := range vh.baseVirtualHost.Locations {
		location := locationBase
		location.Handlers = append([]Handler(nil), location.Handlers...)
		location.Name = match
		if err := json.Unmarshal(locationBuff, &location); err != nil {
			return err
		}
		vh.Locations = append(vh.Locations, &location)
	}
	vh.baseVirtualHost.Locations = nil

	return nil
}

// Validate checks the virtual host config for logical errors.
func (vh *VirtualHost) Validate() error {
	if vh.Name == "" {
		return fmt.Errorf("All virtual hosts should have a name setting")
	}

	return nil
}

func (vh *VirtualHost) String() string {
	return vh.Name
}

// GetSubsections returns the vhost config subsections.
func (vh *VirtualHost) GetSubsections() []Section {
	res := []Section{vh.Logger}
	for _, handler := range vh.Handlers {
		res = append(res, handler)
	}

	for _, l := range vh.Locations {
		res = append(res, l)
	}
	return res
}

func newVHostFromHTTP(h *HTTP) VirtualHost {
	return VirtualHost{parent: h,
		Location: Location{
			baseLocation: baseLocation{
				Handlers:     append([]Handler(nil), h.DefaultHandlers...),
				UpstreamType: h.DefaultUpstreamType,
				CacheZone:    h.DefaultCacheZone,
				Logger:       &h.Logger,
			}}}
}
