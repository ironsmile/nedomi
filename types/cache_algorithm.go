package types

// CacheAlgorithm interface defines how a cache should behave
type CacheAlgorithm interface {

	// Lookup returns wheather this object is in the cache or not
	Lookup(ObjectIndex) bool

	// ShouldKeep is called to signal that this ObjectIndex has been stored
	ShouldKeep(ObjectIndex) bool

	// AddObject adds this ObjectIndex to the cache. Returns an error when
	// the object is in the cache already.
	AddObject(ObjectIndex) error

	// PromoteObject is called every time this part of a file has been used
	// to satisfy a client request
	PromoteObject(ObjectIndex)

	// ConsumedSize returns the full size of all files currently in the cache
	ConsumedSize() BytesSize

	// Stats returns statistics for this cache algorithm
	Stats() CacheStats
}
