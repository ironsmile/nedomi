// Package types describes few of the essential types used throughout the application.
package types

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
)

// ObjectIDHashSize is the size of the byte array that contains the object hash.
const ObjectIDHashSize = sha1.Size

// ObjectIDHash is the fixed-width byte array that represents an ObjectID hash.
type ObjectIDHash [ObjectIDHashSize]byte

// ObjectID represents a cached file.
type ObjectID struct {
	cacheKey string
	path     string
	hash     ObjectIDHash
	//!TODO: add vary headers information
}

func (oid *ObjectID) String() string {
	return fmt.Sprintf("{%x:%s:%s}", oid.Hash(), oid.cacheKey, oid.path)
}

// CacheKey returns the object's cache key.
func (oid *ObjectID) CacheKey() string {
	return oid.cacheKey
}

// Path returns the object's path.
func (oid *ObjectID) Path() string {
	return oid.path
}

// Hash returns the pre-calculated sha1 hash of the object id.
func (oid *ObjectID) Hash() ObjectIDHash {
	return oid.hash
}

// StrHash returns the sha1 hash of the object id in hex format.
func (oid *ObjectID) StrHash() string {
	return hex.EncodeToString(oid.hash[:])
}

// MarshalJSON is used to help the JSON library marshal the unexported vars.
func (oid *ObjectID) MarshalJSON() ([]byte, error) {
	return json.Marshal([]string{oid.cacheKey, oid.path})
}

// UnmarshalJSON is used to help the JSON library unmarshal the unexported vars.
func (oid *ObjectID) UnmarshalJSON(buf []byte) error {
	data := []string{}
	if err := json.Unmarshal(buf, &data); err != nil {
		return err
	}

	if len(data) != 2 || data[0] == "" || data[1] == "" {
		return fmt.Errorf("Invalid ObjectID %s", buf)
	}
	*oid = *NewObjectID(data[0], data[1])
	return nil
}

// NewObjectID creates and returns a new ObjectID.
func NewObjectID(cacheKey, path string) *ObjectID {
	return &ObjectID{
		cacheKey: cacheKey,
		path:     path,
		hash:     sha1.Sum([]byte(cacheKey + "/" + path)),
	}
}
