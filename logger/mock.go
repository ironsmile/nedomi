package logger

import "github.com/ironsmile/nedomi/logger/mock"

// NewMock returns a mock logger instance that can be used for testing.
func NewMock() *mock.Mock {
	l, _ := mock.New(nil)
	return l
}
