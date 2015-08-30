package types

import "io"

// Storage represents a single unit of storage.
type Storage interface {
	// Returns all metadata for this object.
	GetMetadata(id *ObjectID) (*ObjectMetadata, error)

	// Returns an io.ReadCloser that will read the specified part of the object.
	GetPart(id *ObjectIndex) (io.ReadCloser, error)

	// Saves the supplied metadata to the storage.
	SaveMetadata(m *ObjectMetadata) error

	// Saves the contents of the supplied object part to the storage.
	SavePart(index *ObjectIndex, data io.Reader) error

	// Discard an object and its metadata from the storage.
	Discard(id *ObjectID) error

	// Discard the specified part of an Object from the storage.
	DiscardPart(index *ObjectIndex) error

	// Walk iterates over the storage contents. It is used for restoring the
	// state after the service is restarted.
	Iterate(doneCh <-chan struct{}) <-chan *StorageIterObj
}

// StorageIterObj is returned by Iterate()'s channel. It either contains the information
type StorageIterObj struct {
	Error error
	Obj   *ObjectMetadata
	Parts ObjectIndexMap
}
