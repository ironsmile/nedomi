package types

import (
	"io"
	"net/http"

	"golang.org/x/net/context"
)

// StorageOrchestrator is responsible for coordinating and synchronizing the
// operations between the Storage, CacheAlgorithm and each vhost's Upstream server.
type StorageOrchestrator interface {
	Handle(ctx context.Context, req *http.Request) (*ObjectMetadata, io.ReadCloser)
}
