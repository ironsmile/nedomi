package utils

import (
	"bytes"
)

// *NOT* concurrently save error saving a multiple of errors
type CompositeError struct {
	errors []error
}

// returns a string representation of all errors in the order they were appended
func (c *CompositeError) Error() string {
	var b bytes.Buffer
	for ind, err := range c.errors {
		b.WriteString(err.Error())
		if ind == len(c.errors)-1 {
			break
		}
		b.WriteRune('\n')
	}

	return b.String()
}

// append an error
func (c *CompositeError) AppendError(err error) {
	c.errors = append(c.errors, err)
}

func (c *CompositeError) Empty() bool {
	return len(c.errors) == 0
}
