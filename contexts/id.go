package contexts

import (
	"github.com/ironsmile/nedomi/types"
	"golang.org/x/net/context"
)

// The key type is unexported to prevent collisions with context keys defined in
// other packages.
type idContextKey int

const idKey idContextKey = 0

// NewIDContext returns a new Context carrying the supplied Request ID.
func NewIDContext(ctx context.Context, id types.ID) context.Context {
	return context.WithValue(ctx, idKey, id)
}

// GetID extracts the types.ID object, if present.
func GetID(ctx context.Context) (types.ID, bool) {
	id, ok := ctx.Value(idKey).(types.ID)
	return id, ok
}

// AppendToID retruns a new Context carrying an request ID that is
// the once the provided context was carrying with appnded the suplied suffix
func AppendToID(ctx context.Context, suffix []byte) (context.Context, types.ID) {
	var oldID, _ = GetID(ctx)
	var newID = types.ID(append(oldID, suffix...))
	return NewIDContext(ctx, newID), newID
}
