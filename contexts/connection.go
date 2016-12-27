package contexts

import (
	"context"

	"github.com/ironsmile/nedomi/types"
)

// The key type is unexported to prevent collisions with context keys defined in
// other packages.
type connContextKey int

const connKey connContextKey = 0

// NewConnContext returns a new Context carrying the supplied incoming connection.
func NewConnContext(ctx context.Context, conn types.IncomingConn) context.Context {
	return context.WithValue(ctx, connKey, conn)
}

// GetConn extracts the types.IncomingConnection object, if present.
func GetConn(ctx context.Context) (types.IncomingConn, bool) {
	conn, ok := ctx.Value(connKey).(types.IncomingConn)
	return conn, ok
}
