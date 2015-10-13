package proxy

import (
	"fmt"

	"github.com/ironsmile/nedomi/config"
	"github.com/ironsmile/nedomi/types"
)

// Settings contains the possible settings for the proxy
type Settings struct {
}

// New returns a configured and ready to use Upstream instance.
func New(cfg *config.Handler, l *types.Location, next types.RequestHandler) (*ReverseProxy, error) {
	if l.Upstream == nil {
		return nil, fmt.Errorf("No upstream address for proxy handler in %s", l.Name)
	}

	//!TODO: record statistics (times, errors, etc.) for all requests
	return &ReverseProxy{
		Transport: l.Upstream,
		Logger:    l.Logger,
	}, nil
}
