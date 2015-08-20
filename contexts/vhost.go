package contexts

import (
	"github.com/ironsmile/nedomi/types"
	"golang.org/x/net/context"
)

// The key type is unexported to prevent collisions with context keys defined in
// other packages.
type vhostContextKey int

const vKey vhostContextKey = 0

// NewVhostContext returns a new Context carrying the supplied vhost pointer.
func NewVhostContext(ctx context.Context, vhost *types.VirtualHost) context.Context {
	return context.WithValue(ctx, vKey, vhost)
}

// GetVhost extracts the vhost.Vhost object, if present.
func GetVhost(ctx context.Context) (*types.VirtualHost, bool) {
	vhost, ok := ctx.Value(vKey).(*types.VirtualHost)
	return vhost, ok
}
