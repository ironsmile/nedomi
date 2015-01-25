/*
   Package types describes few of the essential types used throughout the application.
*/
package types

// Represents a cached file
type ObjectID string

// Represents particular index in a file
type ObjectIndex struct {
	ObjID ObjectID
	Part  uint32
}

func (o *ObjectID) String() string {
	return (string)(*o)
}
