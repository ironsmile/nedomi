package proxy

import (
	"encoding/json"
	"fmt"

	"github.com/ironsmile/nedomi/config"
	"github.com/ironsmile/nedomi/types"
)

// Settings contains the possible settings for the proxy
type Settings struct {
	UserAgent              string `json:"user_agent"`
	HostHeader             string `json:"host_header"`
	HostHeaderKeepOriginal bool   `json:"host_header_keep_original"`
}

// New returns a configured and ready to use Upstream instance.
func New(cfg *config.Handler, l *types.Location, next types.RequestHandler) (*ReverseProxy, error) {
	if next != nil {
		return nil, types.NotNilNextHandler(cfg.Type)
	}

	if l.Upstream == nil {
		return nil, fmt.Errorf("No upstream set for proxy handler in %s", l.Name)
	}

	s := Settings{
		UserAgent: "nedomi",
	}

	if len(cfg.Settings) != 0 {
		if err := json.Unmarshal(cfg.Settings, &s); err != nil {
			return nil, fmt.Errorf("handler.proxy got error while parsing settings: %s", err)
		}
	}

	//!TODO: record statistics (times, errors, etc.) for all requests

	return &ReverseProxy{
		Upstream: l.Upstream,
		Logger:   l.Logger,
		Settings: s,
	}, nil
}
