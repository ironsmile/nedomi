package utils

import (
	"errors"
	"fmt"
	"runtime"
)

const stackSize = 64 << 10

// NewErrorWithStack returns a new error with the current stack
func NewErrorWithStack(msg string) error {
	return WrapErrorWithStack(errors.New(msg))
}

// WrapErrorWithStack returns an error with the current stack wrapping the provided one
func WrapErrorWithStack(err error) error {
	if err == nil {
		return nil
	}

	buf := make([]byte, stackSize)
	buf = buf[:runtime.Stack(buf, false)]
	return &ErrorWithStack{stack: buf, err: err}
}

// ErrorWithStack is an error which wraps around another error and records the current stack
type ErrorWithStack struct {
	stack []byte
	err   error
}

// Original returns the original error
func (s *ErrorWithStack) Original() error {
	return s.err
}

func (s *ErrorWithStack) Error() string {
	return fmt.Sprintf("%s at \n%s", s.err, s.stack)
}
