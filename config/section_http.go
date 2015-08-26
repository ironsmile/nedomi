package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
)

// BaseHTTP contains the basic configuration options for HTTP.
type BaseHTTP struct {
	Listen         string            `json:"listen"`
	Servers        []json.RawMessage `json:"virtual_hosts"`
	MaxHeadersSize int               `json:"max_headers_size"`
	ReadTimeout    uint32            `json:"read_timeout"`
	WriteTimeout   uint32            `json:"write_timeout"`

	// Defaults for vhosts:
	DefaultHandlerType  string        `json:"default_handler"`
	DefaultUpstreamType string        `json:"default_upstream_type"`
	DefaultCacheZone    string        `json:"default_cache_zone"`
	Logger              LoggerSection `json:"logger"`
}

// HTTP contains all configuration options for HTTP.
type HTTP struct {
	BaseHTTP
	Servers []*VirtualHost `json:"virtual_hosts"`
	parent  *Config
}

// UnmarshalJSON is a custom JSON unmashalling that also implements inheritance,
// custom field initiation and data validation for the HTTP config.
func (h *HTTP) UnmarshalJSON(buff []byte) error {
	if err := json.Unmarshal(buff, &h.BaseHTTP); err != nil {
		return err
	}

	// Inherit HTTP values to vhosts
	baseVhost := VirtualHost{parent: h, BaseVirtualHost: BaseVirtualHost{
		HandlerType:  h.DefaultHandlerType,
		UpstreamType: h.DefaultUpstreamType,
		CacheZone:    h.DefaultCacheZone,
		Logger:       &h.Logger,
	}}

	// Parse all the vhosts
	for _, vhostBuff := range h.BaseHTTP.Servers {
		vhost := baseVhost
		if err := json.Unmarshal(vhostBuff, &vhost); err != nil {
			return err
		}
		h.Servers = append(h.Servers, &vhost)
	}

	h.BaseHTTP.Servers = nil // Cleanup
	return nil
}

// Validate checks the HTTP config for logical errors.
func (h *HTTP) Validate() error {

	if h.Listen == "" {
		return errors.New("Empty `http.listen` directive")
	}

	if len(h.Servers) == 0 {
		return errors.New("There has to be at least one virtual host")
	}

	// Validate that vhosts do not use the same key for the same cache zone
	type czPair struct{ zone, key string }
	usedCzPairs := make(map[czPair]bool)
	for _, vhost := range h.Servers {
		key := czPair{vhost.CacheZone.ID, vhost.CacheKey}
		if usedCzPairs[key] {
			return fmt.Errorf("Virtual host %s has the same cache zone and cache key as another host", vhost.Name)
		}
		usedCzPairs[key] = true
	}

	//!TODO: make sure Listen is valid tcp address
	if _, err := net.ResolveTCPAddr("tcp", h.Listen); err != nil {
		return err
	}

	return nil
}

// GetSubsections returns a slice with all the subsections of the HTTP config.
func (h *HTTP) GetSubsections() []Section {
	res := []Section{h.Logger}
	for _, s := range h.Servers {
		res = append(res, s)
	}
	return res
}
