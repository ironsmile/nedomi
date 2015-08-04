package app

import "golang.org/x/net/context"

// The key type is unexported to prevent collisions with context keys defined in
// other packages.
type applicationContextKey int

const appKey applicationContextKey = 0

// NewContext returns a new Context carrying a pointer to the supplied Application.
func NewContext(ctx context.Context, app *Application) context.Context {
	return context.WithValue(ctx, appKey, app)
}

// FromContext extracts the Application pointer from ctx, if present.
func FromContext(ctx context.Context) (*Application, bool) {
	// ctx.Value returns nil if ctx has no value for the key;
	// the net.IPApplication type assertion returns ok=false for nil.
	app, ok := ctx.Value(appKey).(*Application)
	return app, ok
}
