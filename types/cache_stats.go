/*
   This file contains types for generating the cache statistics page.
*/

package types

import (
	"github.com/ironsmile/nedomi/config"
)

/*
   Every cache should be able to generate a Stats object which is to be used in the
   server status page.
*/
type CacheStats interface {

	// CacheHitPrc returns a string such as '53%' which represents the cache hit
	// ratio of this cache. Basically this number is (Hits()/Requests()) * 100.
	CacheHitPrc() string

	// ID is a way of identifing this particular cache zone
	ID() string

	// Hits returns the number of cache hits this cache has generated
	Hits() uint64

	// Requests returns the number of lookups in the cache
	Requests() uint64

	// Objects returns the number of cache object at the moment. These are the actual
	// objects on the disk. Not maximum possible objects
	Objects() uint64

	// Size returns the consumed space in bytes for this cache
	Size() config.BytesSize
}
