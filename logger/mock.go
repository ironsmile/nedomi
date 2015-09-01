package logger

import (
	"github.com/ironsmile/nedomi/logger/nillogger"
	"github.com/ironsmile/nedomi/types"
)

// NewMock returns a mock logger instance that can be used for testing.
func NewMock() types.Logger {
	//!TODO: do not use the nil logger, write a simple mock logger that remembers
	// the logged things in slices

	l, _ := nillogger.New(nil)
	return l
}
