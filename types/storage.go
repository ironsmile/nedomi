package types

import "io"

// Storage represents a single unit of storage.
type Storage interface {
	// Returns the maximum part size for the storage.
	PartSize() uint64

	// Returns the metadata for this object, it it is present. If the requested
	// metadata is not on the storage, it returns os.ErrNotExist.
	GetMetadata(id *ObjectID) (*ObjectMetadata, error)

	// Returns an io.ReadCloser instance that will read the specified part of
	// the object, if it is present. If the requested part is not on the
	// storage, it will return os.ErrNotExist.
	GetPart(id *ObjectIndex) (io.ReadCloser, error)

	// GetAvailableParts returns types.ObjectIndexMap including all the available
	// parts of for the object specified by the provided objectMetadata
	GetAvailableParts(id *ObjectID) (ObjectIndexMap, error)

	// Saves the supplied metadata to the storage.
	SaveMetadata(m *ObjectMetadata) error

	// Saves the contents of the supplied object part to the storage.
	SavePart(index *ObjectIndex, data io.Reader) error

	// Discard an object and its metadata from the storage.
	Discard(id *ObjectID) error

	// Discard the specified part of an Object from the storage.
	DiscardPart(index *ObjectIndex) error

	// Iterate iterates over the storage objects and passes them and information
	// about their parts to the supplied callback function. It is used for
	// restoring the state after the service has been restarted. When the
	// callback returns false, the iteration stops.
	Iterate(callback func(*ObjectMetadata, ObjectIndexMap) bool) error
}

//!TODO: use custom error type instead of os.ErrNotExist?
