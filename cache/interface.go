// Package cache is implements the caching algorithm. It defines the Manager
// interface. Every CacheZone has its own cache manager. This makes it possible for
// different caching algorithms to be used in the same time.
package cache

import "github.com/ironsmile/nedomi/types"

// Manager interface defines how a cache should behave
type Manager interface {

	// Lookup returns wheather this object is in the cache or not
	Lookup(types.ObjectIndex) bool

	// ShouldKeep is called to signal that this ObjectIndex has been stored
	ShouldKeep(types.ObjectIndex) bool

	// AddObject adds this ObjectIndex to the cache. Returns an error when
	// the object is in the cache already.
	AddObject(types.ObjectIndex) error

	// PromoteObject is called every time this part of a file has been used
	// to satisfy a client request
	PromoteObject(types.ObjectIndex)

	// ConsumedSize returns the full size of all files currently in the cache
	ConsumedSize() types.BytesSize

	// ReplaceRemoveChannel makes this cache communicate its desire to remove objects
	// on this channel
	ReplaceRemoveChannel(chan<- types.ObjectIndex)

	// Stats returns statistics for this cache manager
	Stats() types.CacheStats
}
