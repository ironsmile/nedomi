package cache

import (
	"fmt"
	"net/http"

	"github.com/ironsmile/nedomi/config"
	"github.com/ironsmile/nedomi/types"
	"golang.org/x/net/context"
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
	//!TODO: remove the need for "upstream" and make it the `next` RequestHandler
	if loc.Upstream == nil {
		return nil, fmt.Errorf("Caching proxy handler for %s needs a working upstream.", loc.Name)
	}

	if loc.Cache.Storage == nil || loc.Cache.Algorithm == nil {
		return nil, fmt.Errorf("Caching proxy handler for %s needs a configured cache zone.", loc.Name)
	}

	return &CachingProxy{loc, cfg, next}, nil
}

// RequestHandle is the main serving function
func (c *CachingProxy) RequestHandle(ctx context.Context,
	resp http.ResponseWriter, req *http.Request, _ *types.Location) {

	objID := types.NewObjectID(c.CacheKey, req.URL.String())
	rh := &reqHandler{c, ctx, req, resp, objID, nil}
	rh.handle()
}
