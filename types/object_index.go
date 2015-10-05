// Package types describes few of the essential types used throughout the application.
package types

import (
	"encoding/binary"
	"fmt"
)

// ObjectIndexHashSize is the size of the byte array that contains the object hash.
const ObjectIndexHashSize = ObjectIDHashSize + 4 // sizeof(uint32)

// ObjectIndexHash is the fixed-width byte array that represents an ObjectID hash.
type ObjectIndexHash [ObjectIndexHashSize]byte

// ObjectIndex represents a particular index in a file
type ObjectIndex struct {
	ObjID *ObjectID
	Part  uint32
	hash  ObjectIndexHash
}

func (oi *ObjectIndex) String() string {
	return fmt.Sprintf("%s:%d", oi.ObjID, oi.Part)
}

// Hash returns the pre-calculated sha1 of the ObjectID with concatinated the 4 bytes identifying the part
func (oi *ObjectIndex) Hash() ObjectIndexHash {
	var hash ObjectIndexHash
	copy(hash[:], oi.ObjID.hash[:])
	binary.BigEndian.PutUint32(hash[ObjectIDHashSize:], oi.Part)
	return hash
}

// HashStr returns a hash
func (oi *ObjectIndex) HashStr() string {
	return fmt.Sprintf("%s:%d", oi.ObjID.StrHash(), oi.Part)
}
