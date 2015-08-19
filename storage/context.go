package storage

import "golang.org/x/net/context"

// The key type is unexported to prevent collisions with context keys defined in
// other packages.
type storageContextKey int

const sKey storageContextKey = 0

// NewContext returns a new Context carrying the map with the supplied storages.
func NewContext(ctx context.Context, storages map[string]Storage) context.Context {
	return context.WithValue(ctx, sKey, storages)
}

// FromContext extracts the map of storage.Storage objects, if present.
func FromContext(ctx context.Context) (map[string]Storage, bool) {
	// ctx.Value returns nil if ctx has no value for the key;
	// the map[string]storage.Storage type assertion returns ok=false for nil.
	storages, ok := ctx.Value(sKey).(map[string]Storage)
	return storages, ok
}
