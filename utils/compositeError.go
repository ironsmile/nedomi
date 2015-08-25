package utils

import (
	"bytes"
)

// CompositeError is used for saving multiple errors.
// IMPORTANT: *NOT* safe for concurrent usage
type CompositeError []error

// Returns a string representation of all errors in the order they were appended.
func (c *CompositeError) Error() string {
	var b bytes.Buffer
	for ind, err := range *c {
		b.WriteString(err.Error())
		if ind == len(*c)-1 {
			break
		}
		b.WriteRune('\n')
	}

	return b.String()
}

// AppendError is used for adding another error to the list.
func (c *CompositeError) AppendError(err error) {
	if err != nil {
		*c = append(*c, err)
	}
}

// Empty returns true if the internal error list is empty.
func (c *CompositeError) Empty() bool {
	return len(*c) == 0
}
