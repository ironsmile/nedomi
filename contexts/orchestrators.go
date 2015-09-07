package contexts

import (
	"github.com/ironsmile/nedomi/types"
	"golang.org/x/net/context"
)

// The key type is unexported to prevent collisions with context keys defined in
// other packages.
type orchestratorContextKey int

const sKey orchestratorContextKey = 0

// NewStorageOrchestratorsContext returns a new Context carrying the map with
// the supplied orchestrators.
func NewStorageOrchestratorsContext(ctx context.Context,
	orchestrators map[string]types.StorageOrchestrator) context.Context {

	return context.WithValue(ctx, sKey, orchestrators)
}

// GetStorageOrchestrators extracts the map of types.Orchestrator objects, if present.
func GetStorageOrchestrators(ctx context.Context) (map[string]types.StorageOrchestrator, bool) {
	// ctx.Value returns nil if ctx has no value for the key;
	// the map[string]types.StorageOrchestrator type assertion returns ok=false for nil.
	orchestrators, ok := ctx.Value(sKey).(map[string]types.StorageOrchestrator)
	return orchestrators, ok
}
