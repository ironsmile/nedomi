package config

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

// Upstream contains all configuration options for an upstream group.
type Upstream struct {
	ID        string
	Balancing string            `json:"balancing"`
	Addresses []UpstreamAddress `json:"addresses"`
	Settings  UpstreamSettings  `json:"settings"`
}

// UpstreamSettings contains all possible upstream settings.
type UpstreamSettings struct {
	MaxConnectionsPerServer uint32 `json:"max_connections_per_server"`
	UseIPv4                 bool   `json:"use_ipv4"`
	UseIPv6                 bool   `json:"use_ipv6"`
	ResolveAddresses        bool   `json:"resolve_addresses"`
	//!TODO: add settings for timeouts, keep-alives, retries, etc.
}

// UpstreamAddress contains a single upstream URL and it's weight.
type UpstreamAddress struct {
	URL    *url.URL
	Weight uint32
}

// DefaultUpstreamWeight is the weight that is assigned to upstreams with no
// specified weight.
const DefaultUpstreamWeight uint32 = 100

// Validate checks a CacheZone config section for errors.
func (cz *Upstream) Validate() error {
	if len(cz.Addresses) < 1 {
		return fmt.Errorf("Upstream %s has no addresses!", cz.ID)
	}

	return nil
}

// GetSubsections returns nil (CacheZone has no subsections).
func (cz *Upstream) GetSubsections() []Section {
	return nil
}

// UnmarshalJSON is a custom JSON unmarshalling which parses upstream addresses
// in the format "http://some.url|weight%"
func (addr *UpstreamAddress) UnmarshalJSON(buff []byte) error {
	val := strings.Trim(string(buff), "\"")
	if len(val) == 0 {
		return fmt.Errorf("Invalid upstream address '%s'", buff)
	}
	data := strings.SplitN(val, "|", 2)

	parsed, err := url.Parse(data[0])
	if err != nil {
		return fmt.Errorf("Error upstream address %s: %s", data[0], err)
	}
	addr.URL = parsed

	// If there is no weight, assign the DefaultUpstreamWeight
	if len(data) == 1 {
		addr.Weight = DefaultUpstreamWeight
		return nil
	}

	w, err := strconv.ParseUint(data[1], 10, 32)
	if err != nil {
		return err
	}
	addr.Weight = uint32(w)
	return nil
}

// GetDefaultUpstreamSettings returns some sane dafault settings for upstreams
func GetDefaultUpstreamSettings() UpstreamSettings {
	return UpstreamSettings{
		MaxConnectionsPerServer: 0, // Unlimited connection number by default
		UseIPv4:                 true,
		UseIPv6:                 false,
		ResolveAddresses:        true,
		//!TODO: add settings for timeouts, keep-alives, retries, etc.
	}
}
