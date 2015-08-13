package cache

import "golang.org/x/net/context"

// The key type is unexported to prevent collisions with context keys defined in
// other packages.
type cacheAlgorithmsContextKey int

const cmKey cacheAlgorithmsContextKey = 0

// NewContext returns a new Context carrying a slice of Cache
func NewContext(ctx context.Context, cms map[string]Algorithm) context.Context {
	return context.WithValue(ctx, cmKey, cms)
}

// FromContext extracts the slice of cache.Algorithm objects, if present.
func FromContext(ctx context.Context) (map[string]Algorithm, bool) {
	// ctx.Value returns nil if ctx has no value for the key;
	// the map[string]cache.Algorithm type assertion returns ok=false for nil.
	cms, ok := ctx.Value(cmKey).(map[string]Algorithm)
	return cms, ok
}
