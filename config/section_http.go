package config

import "encoding/json"

// HTTPSectionBase contains the basic configuration options for HTTP.
type HTTPSectionBase struct {
	Listen         string            `json:"listen"`
	Servers        []json.RawMessage `json:"virtual_hosts"`
	MaxHeadersSize int               `json:"max_headers_size"`
	ReadTimeout    uint32            `json:"read_timeout"`
	WriteTimeout   uint32            `json:"write_timeout"`

	// Defaults for vhosts:
	//!TODO: rename all default vars to be like "default_sth" or "DefaultSth"
	CacheAlgo    string        `json:"cache_algorithm"`
	UpstreamType string        `json:"upstream_type"`
	HandlerType  string        `json:"handler"`
	Logger       LoggerSection `json:"logger"`
}

// HTTPSection contains all configuration options for HTTP.
type HTTPSection struct {
	HTTPSectionBase
	Servers []*VirtualHost `json:"virtual_hosts"`
	parent  *Config
}

// UnmarshalJSON is a custom JSON unmashalling that also implements inheritance,
// custom field initiation and data validation for the HTTP config.
func (h *HTTPSection) UnmarshalJSON(buff []byte) error {
	if err := json.Unmarshal(buff, &h.HTTPSectionBase); err != nil {
		return err
	}

	// Inherit HTTP values to vhosts
	baseVhost := VirtualHost{parent: h, VirtualHostBase: VirtualHostBase{
		CacheAlgo:    h.CacheAlgo,
		HandlerType:  h.HandlerType,
		UpstreamType: h.UpstreamType,
		Logger:       &h.Logger}}

	// Parse all the vhosts
	for _, vhostBuff := range h.HTTPSectionBase.Servers {
		vhost := baseVhost
		if err := json.Unmarshal(vhostBuff, &vhost); err != nil {
			return err
		}
		h.Servers = append(h.Servers, &vhost)
	}

	h.HTTPSectionBase.Servers = nil // Cleanup
	return h.Validate()
}

// Validate checks the HTTP config for logical errors.
func (h *HTTPSection) Validate() error {
	//!TODO: implement
	return nil
}
