package types

import "fmt"

// NilNextHandler is an error type to be returned when a next handler is required but it's nil
type NilNextHandler string

func (n NilNextHandler) Error() string {
	return fmt.Sprintf("%s: next handler is required but it's nil", n)
}

// NotNilNextHandler is an error type to be returned when a next handler is not nil but it should be
type NotNilNextHandler string

func (n NotNilNextHandler) Error() string {
	return fmt.Sprintf("%s: next handler is not nil but it should be", n)
}
