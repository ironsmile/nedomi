package cache

import (
	"github.com/gophergala/nedomi/types"
)

/*
   CacheManager interface defines how a cache should behave
*/
type CacheManager interface {

	// Has returns wheather this object is in the cache or not
	Has(types.ObjectIndex) bool

	// ObjectIndexStored is called when
	ObjectIndexStored(types.ObjectIndex) error

	// AddObjectIndex adds this ObjectIndex to the cache
	AddObjectIndex(types.ObjectIndex) error
}
