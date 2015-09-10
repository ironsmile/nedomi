package types

import (
	"net/http"

	"golang.org/x/net/context"
)

// StorageOrchestrator is responsible for coordinating and synchronizing the
// operations between the Storage, CacheAlgorithm and each vhost's Upstream server.
type StorageOrchestrator interface {

	// This method handles client requests for which the cache can be used (the
	// request does not have no-cache headers).
	Handle(context.Context, http.ResponseWriter, *http.Request, *Location)

	// Returns the cache statistics.
	GetCacheStats() CacheStats
}
