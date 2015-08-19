// Package types describes few of the essential types used throughout the application.
package types

import "fmt"

// ObjectID represents a cached file
type ObjectID struct {
	CacheKey string
	Path     string
}

func (oid ObjectID) String() string {
	return fmt.Sprintf("%s:%s", oid.CacheKey, oid.Path)
}
