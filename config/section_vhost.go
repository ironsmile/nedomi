package config

import (
	"encoding/json"
	"fmt"
	"net/url"
)

// BaseVirtualHost contains the basic configuration options for virtual hosts.
type BaseVirtualHost struct {
	Locations map[string]json.RawMessage `json:"locations"`
}

// VirtualHost contains all configuration options for virtual hosts. It
// redefines some of the base fields to use the correct types.
type VirtualHost struct {
	BaseVirtualHost
	LocationSection
	Locations []*LocationSection `json:"locations"`
	parent    *HTTP
}

// UnmarshalJSON is a custom JSON unmashalling that also implements inheritance
// and custom field initiation
func (vh *VirtualHost) UnmarshalJSON(buff []byte) error {
	// Parse the base values
	if err := json.Unmarshal(buff, &vh.BaseLocationSection); err != nil {
		return err
	}
	if err := json.Unmarshal(buff, &vh.BaseVirtualHost); err != nil {
		return err
	}

	// Convert the upstream URL from string to url.URL
	if vh.BaseLocationSection.UpstreamAddress != "" {
		parsed, err := url.Parse(vh.BaseLocationSection.UpstreamAddress)
		if err != nil {
			return fmt.Errorf("Error parsing server %s upstream. %s", vh.Name, err)
		}
		vh.UpstreamAddress = parsed
	}

	// Inject the cache zone configuration from the root config
	vh.CacheZone = vh.parent.parent.CacheZones[vh.BaseLocationSection.CacheZone]

	baseLocation := LocationSection{
		parent: vh,
		BaseLocationSection: BaseLocationSection{
			HandlerType:     vh.HandlerType,
			UpstreamType:    vh.UpstreamType,
			UpstreamAddress: vh.BaseLocationSection.UpstreamAddress,
			CacheZone:       vh.BaseLocationSection.CacheZone,
			CacheKey:        vh.BaseLocationSection.CacheKey,
			Logger:          vh.Logger,
		},
	}

	// Parse all the locations
	for match, locationBuff := range vh.BaseVirtualHost.Locations {
		location := baseLocation
		location.Name = match
		if err := json.Unmarshal(locationBuff, &location); err != nil {
			return err
		}
		vh.Locations = append(vh.Locations, &location)
	}

	return nil
}

// Validate checks the virtual host config for logical errors.
func (vh *VirtualHost) Validate() error {
	if vh.Name == "" {
		return fmt.Errorf("All virtual hosts should have a name setting")
	}

	return nil
}

// GetSubsections returns the vhost config subsections.
func (vh *VirtualHost) GetSubsections() []Section {
	res := []Section{vh.Logger, vh.HandlerType}

	for _, l := range vh.Locations {
		res = append(res, l)
	}
	return res
}

func newVHostFromHTTP(h *HTTP) VirtualHost {
	return VirtualHost{parent: h,
		LocationSection: LocationSection{
			BaseLocationSection: BaseLocationSection{
				HandlerType:  h.DefaultHandlerType,
				UpstreamType: h.DefaultUpstreamType,
				CacheZone:    h.DefaultCacheZone,
				Logger:       &h.Logger,
			}}}
}
