package storage

import (
	"io"

	. "github.com/gophergala/nedomi/types"
)

// A unit of Storage
type Storage interface {
	// Returns a io.ReadCloser that will read from the `start`
	// of an object with ObjectId `id` to the `end`.
	Get(id ObjectID, start, end int64) (io.ReadCloser, error)
	// Discard an object from the storage
	Discard(id ObjectID) error
	// Discard an index of an Object from the storage
	DiscardIndex(index ObjectIndex) error
}
