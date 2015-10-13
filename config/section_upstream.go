package config

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

// Upstream contains all configuration options for an upstream group.
type Upstream struct {
	ID                      string
	Balancing               string            `json:"balancing"`
	MaxConnectionsPerServer uint32            `json:"max_connections_per_server"`
	Addresses               []UpstreamAddress `json:"addresses"`
}

// UpstreamAddress contains a single upstream URL and it's weight
type UpstreamAddress struct {
	URL *url.URL

	// Weight is the percentage of requests that this upsream address should
	// handle. 0 means that the percent should be calculated based on the other
	// weights.
	Weight float64
}

// Validate checks a CacheZone config section for errors.
func (cz *Upstream) Validate() error {
	if len(cz.Addresses) < 1 {
		return fmt.Errorf("Upstream %s has no addresses!", cz.ID)
	}

	var pSum float64
	for _, addr := range cz.Addresses {
		pSum += addr.Weight
	}
	if pSum > 1 {
		return fmt.Errorf("Upstream %s has addresses with total weight above 100", cz.ID)
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

	// If there is a weight percentage, convert it to float
	if len(data) > 1 {
		i := strings.Index(data[1], "%")
		if i < 0 {
			return fmt.Errorf("Percentage sign not found for the weight of upstream address %s", data[0])
		}
		f, err := strconv.ParseFloat(data[1][:i], 64)
		if err != nil {
			return err
		}
		if f <= 0 {
			return fmt.Errorf("Invalid weight percentage %f", f)
		}
		addr.Weight = f / 100
	}

	return nil
}
