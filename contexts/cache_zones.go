package contexts

import (
	"github.com/ironsmile/nedomi/types"
	"golang.org/x/net/context"
)

// The key type is unexported to prevent collisions with context keys defined in
// other packages.
type cacheZoneContextKey int

const sKey cacheZoneContextKey = 0

// NewCacheZonesContext returns a new Context carrying the map with
// the supplied CacheZones.
func NewCacheZonesContext(ctx context.Context,
	cacheZones map[string]types.CacheZone) context.Context {

	return context.WithValue(ctx, sKey, cacheZones)
}

// GetCacheZones extracts the map of types.CacheZone objects, if present.
func GetCacheZones(ctx context.Context) (map[string]types.CacheZone, bool) {
	// ctx.Value returns nil if ctx has no value for the key;
	// the map[string]types.CacheZone type assertion returns ok=false for nil.
	cacheZones, ok := ctx.Value(sKey).(map[string]types.CacheZone)
	return cacheZones, ok
}
