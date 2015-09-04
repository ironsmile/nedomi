package logger

import (
	"github.com/ironsmile/nedomi/logger/buffers"
	"github.com/ironsmile/nedomi/types"
)

// NewMock returns a mock logger instance that can be used for testing.
func NewMock() types.Logger {
	l, _ := buffers.New(nil)
	return l
}
