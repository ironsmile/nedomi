// Package types describes few of the essential types used throughout the application.
package types

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
)

//!TODO: maybe make CacheKey and Path private and add getters for them? It would
// prevent changes in their value after the hash has already been calculated. If
// we want to be albe to change them, we can add setters that reset the hash.

// ObjectID represents a cached file
type ObjectID struct {
	CacheKey         string
	Path             string
	isHashCalculated bool
	hash             [sha1.Size]byte
	//!TODO: add vary headers information
}

func (oid *ObjectID) String() string {
	return fmt.Sprintf("%x:%s:%s", oid.Hash(), oid.CacheKey, oid.Path)
}

// StrHash returns the sha1 hash of the selected object id in hex format.
func (oid *ObjectID) StrHash() string {
	return hex.EncodeToString(oid.Hash())
}

// Hash returns the sha1 hash of the selected object id.
func (oid *ObjectID) Hash() []byte {
	if !oid.isHashCalculated {
		h := sha1.New()
		io.WriteString(h, oid.CacheKey+"/"+oid.Path)
		copy(oid.hash[:], h.Sum(nil))
		oid.isHashCalculated = true
	}
	return oid.hash[:]
}
