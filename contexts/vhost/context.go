package vhost

import (
	"github.com/ironsmile/nedomi/types"
	"golang.org/x/net/context"
)

// The key type is unexported to prevent collisions with context keys defined in
// other packages.
type vhostContextKey int

const vKey vhostContextKey = 0

// NewContext returns a new Context carrying the supplied vhost pointer.
func NewContext(ctx context.Context, vhost *types.VirtualHost) context.Context {
	return context.WithValue(ctx, vKey, vhost)
}

// FromContext extracts the vhost.Vhost object, if present.
func FromContext(ctx context.Context) (*types.VirtualHost, bool) {
	vhost, ok := ctx.Value(vKey).(*types.VirtualHost)
	return vhost, ok
}
