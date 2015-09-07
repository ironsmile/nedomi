package nillogger

import (
	"github.com/ironsmile/nedomi/config"
)

// New returns a new Nil logger.
func New(cfg *config.LoggerSection) (*Nil, error) {
	return &Nil{}, nil
}

// Nil is no configuration noop implementation of the Logger interface.
type Nil struct{}

// Log is a noop
func (n *Nil) Log(v ...interface{}) {
	return
}

// Logf is a noop
func (n *Nil) Logf(format string, args ...interface{}) {
	return
}

// Debug is a noop
func (n *Nil) Debug(v ...interface{}) {
	return
}

// Debugf is a noop
func (n *Nil) Debugf(format string, args ...interface{}) {
	return
}

// Error is a noop
func (n *Nil) Error(v ...interface{}) {
	return
}

// Errorf is a noop
func (n *Nil) Errorf(format string, args ...interface{}) {
	return
}

// Fatal is a noop
func (n *Nil) Fatal(v ...interface{}) {
	return
}

// Fatalf is a noop
func (n *Nil) Fatalf(format string, v ...interface{}) {
	return
}
