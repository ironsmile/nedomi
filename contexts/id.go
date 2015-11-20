package contexts

import (
	"github.com/ironsmile/nedomi/types"
	"golang.org/x/net/context"
)

// The key type is unexported to prevent collisions with context keys defined in
// other packages.
type reqIDContextKey int

const reqIDKey reqIDContextKey = 0

// NewIDContext returns a new Context carrying the supplied Request ID.
func NewIDContext(ctx context.Context, reqID types.RequestID) context.Context {
	return context.WithValue(ctx, reqIDKey, reqID)
}

// GetRequestID extracts the types.RequestID object, if present.
func GetRequestID(ctx context.Context) (types.RequestID, bool) {
	reqID, ok := ctx.Value(reqIDKey).(types.RequestID)
	return reqID, ok
}

// AppendToRequestID retruns a new Context carrying an request ID that is
// the once the provided context was carrying with appnded the suplied suffix
func AppendToRequestID(ctx context.Context, suffix []byte) (context.Context, types.RequestID) {
	var oldReqID, _ = GetRequestID(ctx)
	var newReqID = types.RequestID(append(oldReqID, suffix...))
	return NewIDContext(ctx, newReqID), newReqID
}
