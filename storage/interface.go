/*
	Package storage deals with files on the disk or whatever storage. It defines the
	Storage interface. It has methods for getting contents of a file, headers of a
	file and methods for removing files. Since every cache zone has its own storage
	it is possible to have different storage implementations running at the same
	time.
*/
package storage

import (
	"io"
	"net/http"

	"github.com/gophergala/nedomi/config"
	. "github.com/gophergala/nedomi/types"
)

// A unit of Storage
type Storage interface {
	// Returns a io.ReadCloser that will read from the `start`
	// of an object with ObjectId `id` to the `end`.
	Get(vh *config.VirtualHost, id ObjectID, start, end uint64) (io.ReadCloser, error)

	// Returns a io.ReadCloser that will read the whole file
	GetFullFile(vh *config.VirtualHost, id ObjectID) (io.ReadCloser, error)

	// Returns all headers for this object
	Headers(vh *config.VirtualHost, id ObjectID) (http.Header, error)

	// Discard an object from the storage
	Discard(id ObjectID) error

	// Discard an index of an Object from the storage
	DiscardIndex(index ObjectIndex) error
}
