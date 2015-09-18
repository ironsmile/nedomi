package utils

import "bytes"

// CompositeError is used for saving multiple errors.
// IMPORTANT: *NOT* safe for concurrent usage
type CompositeError []error

// Returns a string representation of all errors in the order they were appended.
func (c *CompositeError) Error() string {
	var b bytes.Buffer
	for ind, err := range *c {
		_, _ = b.WriteString(err.Error())
		if ind == len(*c)-1 {
			break
		}
		_, _ = b.WriteRune('\n')
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
	return c == nil || len(*c) == 0
}

// NewCompositeError returns a new CompositeError with the supplied errors. The
// error interface is returned because of this: https://golang.org/doc/faq#nil_error
func NewCompositeError(errors ...error) error {
	res := &CompositeError{}
	for _, err := range errors {
		res.AppendError(err)
	}
	if res.Empty() {
		return nil
	}
	return res
}
