package types

import "errors"

// CacheAlgorithm interface defines how a cache should behave
type CacheAlgorithm interface {

	// Lookup returns wheather this object is in the cache or not
	Lookup(*ObjectIndex) bool

	// ShouldKeep is called to signal that this ObjectIndex has been stored
	ShouldKeep(*ObjectIndex) bool

	// AddObject adds this ObjectIndex to the cache. Returns an error when
	// the object is in the cache already.
	AddObject(*ObjectIndex) error

	// PromoteObject is called every time this part of a file has been used
	// to satisfy a client request
	PromoteObject(*ObjectIndex)

	// ConsumedSize returns the full size of all files currently in the cache
	ConsumedSize() BytesSize

	// Stats returns statistics for this cache algorithm
	Stats() CacheStats

	// Remove all of the provided object indexes from the cache.
	Remove(...*ObjectIndex)

	// ChangeConfig changes the changeable parts of the a CacheAlgorithm:
	// the timeout and count for removing objects in bulk
	// and the count of objects it contains. Automatically resizing the algorithm
	// if it's required
	ChangeConfig(bulkTimeout, bulkCount, objectCount uint64)

	// SetLogger changes the Logger of the CacheAlgorithm
	SetLogger(Logger)
}

// Exported errors
var (
	ErrAlreadyInCache = errors.New("Object already in cache")
)
