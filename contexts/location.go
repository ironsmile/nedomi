package contexts

import (
	"github.com/ironsmile/nedomi/types"
	"golang.org/x/net/context"
)

// The key type is unexported to prevent collisions with context keys defined in
// other packages.
type locationContextKey int

const lKey locationContextKey = 0

// NewLocationContext returns a new Context carrying the supplied location pointer.
func NewLocationContext(ctx context.Context, l *types.Location) context.Context {
	return context.WithValue(ctx, lKey, l)
}

// GetLocation extracts the types.Locaiton object, if present.
func GetLocation(ctx context.Context) (*types.Location, bool) {
	l, ok := ctx.Value(lKey).(*types.Location)
	return l, ok
}
