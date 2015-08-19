// Package types describes few of the essential types used throughout the application.
package types

import (
	"fmt"
)

// ObjectIndex represents a particular index in a file
type ObjectIndex struct {
	ObjID ObjectID
	Part  uint32
}

func (oi ObjectIndex) String() string {
	return fmt.Sprintf("%s:%d", oi.ObjID, oi.Part)
}
