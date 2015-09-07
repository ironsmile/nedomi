package types

import (
	"io"
	"net/http"

	"golang.org/x/net/context"
)

// StorageOrchestrator is responsible for coordinating and synchronizing the
// operations between the Storage, CacheAlgorithm and each vhost's Upstream server.
type StorageOrchestrator interface {

	// This method handles client requests for which the cache can be used (the
	// request does not have no-cache headers).
	Handle(ctx context.Context, req *http.Request) (*ObjectMetadata, io.ReadCloser, error)

	// Returns the cache statistics.
	GetCacheStats() CacheStats
}
