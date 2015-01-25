/*
   Package types describes few of the essential types used throughout the application.
*/
package types

import (
	"fmt"
)

// Represents a cached file
type ObjectID struct {
	CacheKey string
	Path     string
}

// Represents particular index in a file
type ObjectIndex struct {
	ObjID ObjectID
	Part  uint32
}

func (oid *ObjectID) String() string {
	return fmt.Sprintf("%s:%s", oid.CacheKey, oid.Path)
}
