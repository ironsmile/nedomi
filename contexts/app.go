package contexts

import (
	"context"

	"github.com/ironsmile/nedomi/types"
)

// The key type is unexported to prevent collisions with context keys defined in
// other packages.
type appContextKey int

const aKey appContextKey = 0

// NewAppContext returns a new Context carrying the supplied App.
func NewAppContext(ctx context.Context,
	app types.App) context.Context {

	return context.WithValue(ctx, aKey, app)
}

// GetApp extracts the types.App object, if present.
func GetApp(ctx context.Context) (types.App, bool) {
	// ctx.Value returns nil if ctx has no value for the key;
	// the map[string]types.CacheZone type assertion returns ok=false for nil.
	app, ok := ctx.Value(aKey).(types.App)
	return app, ok
}
