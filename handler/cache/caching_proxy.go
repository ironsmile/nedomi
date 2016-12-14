package cache

import (
	"fmt"
	"net/http"

	"golang.org/x/net/context"

	"github.com/ironsmile/nedomi/config"
	"github.com/ironsmile/nedomi/types"
)

// CachingProxy is resposible for caching the metadata and parts the requested
// objects to `loc.Storage`, according to the `loc.Algorithm`.
type CachingProxy struct {
	*types.Location
	cfg  *config.Handler
	next types.RequestHandler
}

// New creates and returns a ready to used Handler.
func New(cfg *config.Handler, loc *types.Location, next types.RequestHandler) (*CachingProxy, error) {
	if next == nil {
		return nil, fmt.Errorf("caching proxy handler for %s needs a next handler", loc.Name)
	}

	if loc.Cache.Storage == nil || loc.Cache.Algorithm == nil {
		return nil, fmt.Errorf("caching proxy handler for %s needs a configured cache zone", loc.Name)
	}

	return &CachingProxy{loc, cfg, next}, nil
}

// RequestHandle is the main serving function
func (c *CachingProxy) RequestHandle(ctx context.Context, resp http.ResponseWriter, req *http.Request) {
	if req.Method != "GET" && req.Method != "HEAD" {
		c.next.RequestHandle(ctx, resp, req)
		return
	}

	rh := &reqHandler{
		CachingProxy: c,
		ctx:          ctx,
		req:          req,
		resp:         resp,
	}
	rh.handle()
}
