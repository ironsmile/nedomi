package cache

import (
	"fmt"
	"net/http"

	"github.com/ironsmile/nedomi/config"
	"github.com/ironsmile/nedomi/types"
)

// CachingProxy is resposible for caching the metadata and parts the requested
// objects to `loc.Storage`, according to the `loc.Algorithm`.
type CachingProxy struct {
	*types.Location
	cfg  *config.Handler
	next http.Handler
}

// New creates and returns a ready to used Handler.
func New(cfg *config.Handler, loc *types.Location, next http.Handler) (*CachingProxy, error) {
	if next == nil {
		return nil, fmt.Errorf("caching proxy handler for %s needs a next handler", loc.Name)
	}

	if loc.Cache.Storage == nil || loc.Cache.Algorithm == nil {
		return nil, fmt.Errorf("caching proxy handler for %s needs a configured cache zone", loc.Name)
	}

	return &CachingProxy{loc, cfg, next}, nil
}

// ServeHTTP is the main serving function
func (c *CachingProxy) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	if req.Method != "GET" && req.Method != "HEAD" {
		c.next.ServeHTTP(resp, req)
		return
	}

	rh := &reqHandler{
		CachingProxy: c,
		req:          req,
		resp:         resp,
	}
	rh.handle()
}
