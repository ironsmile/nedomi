package types

import (
	"io"

	"golang.org/x/net/context"
)

// Storage represents a single unit of storage.
type Storage interface {
	// Returns all metadata for this object.
	GetMetadata(ctx context.Context, id *ObjectID) (*ObjectMetadata, error)

	// Returns an io.ReadCloser that will read from the `start` of an object
	// with ObjectId `id` to the `end`.
	Get(ctx context.Context, id *ObjectID, start, end uint64) (io.ReadCloser, error)

	// Returns an io.ReadCloser that will read the specified part of the object.
	GetPart(ctx context.Context, id *ObjectIndex) (io.ReadCloser, error)

	// Saves the supplied metadata to the storage.
	SaveMetadata(m *ObjectMetadata) error

	// Saves the contents of the supplied object part to the storage.
	SavePart(index *ObjectIndex, data []byte) error

	// Discard an object and its metadata from the storage.
	Discard(id *ObjectID) error

	// Discard the specified part of an Object from the storage.
	DiscardPart(index *ObjectIndex) error

	// Walk iterates over the storage contents. It is used for restoring the
	// state after the service is restarted.
	Walk() <-chan *ObjectFullMetadata
}
