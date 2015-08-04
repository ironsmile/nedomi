package cache

import "golang.org/x/net/context"

// The key type is unexported to prevent collisions with context keys defined in
// other packages.
type cacheManagersContextKey int

const cmKey cacheManagersContextKey = 0

// NewContext returns a new Context carrying a slice of Cache
func NewContext(ctx context.Context, cms map[uint32]Manager) context.Context {
	return context.WithValue(ctx, cmKey, cms)
}

// FromContext extracts the slice of cache.Manager objects, if present.
func FromContext(ctx context.Context) (map[uint32]Manager, bool) {
	// ctx.Value returns nil if ctx has no value for the key;
	// the map[uint32]cache.Manager type assertion returns ok=false for nil.
	cms, ok := ctx.Value(cmKey).(map[uint32]Manager)
	return cms, ok
}
