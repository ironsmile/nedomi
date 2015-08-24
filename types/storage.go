package types

import (
	"io"
	"net/http"

	"golang.org/x/net/context"
)

// Storage represents a single unit of storage.
type Storage interface {
	// Returns a io.ReadCloser that will read from the `start`
	// of an object with ObjectId `id` to the `end`.
	Get(ctx context.Context, id ObjectID, start, end uint64) (io.ReadCloser, error)

	// Returns a io.ReadCloser that will read the whole file
	GetFullFile(ctx context.Context, id ObjectID) (io.ReadCloser, error)

	// Returns all headers for this object
	Headers(ctx context.Context, id ObjectID) (http.Header, error)

	// Discard an object from the storage
	Discard(id ObjectID) error

	// Discard an index of an Object from the storage
	DiscardIndex(index ObjectIndex) error

	// Returns the used cache algorithm
	GetCacheAlgorithm() *CacheAlgorithm

	// Close the Storage. All calls to this storage instance after this one
	// have an undefined behaviour.
	Close() error
}
