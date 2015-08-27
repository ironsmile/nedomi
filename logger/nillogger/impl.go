package nillogger

import (
	"github.com/ironsmile/nedomi/config"
)

// New returns a new logger that does nothing.
func New(cfg *config.LoggerSection) (*nilLogger, error) {
	return &nilLogger{}, nil
}

type nilLogger struct{}

func (n *nilLogger) Log(v ...interface{}) {
	return
}
func (n *nilLogger) Logf(format string, args ...interface{}) {
	return
}

func (n *nilLogger) Logln(v ...interface{}) {
	return
}

func (n *nilLogger) Debug(v ...interface{}) {
	return
}

func (n *nilLogger) Debugf(format string, args ...interface{}) {
	return
}

func (n *nilLogger) Debugln(v ...interface{}) {
	return
}

func (n *nilLogger) Error(v ...interface{}) {
	return
}

func (n *nilLogger) Errorf(format string, args ...interface{}) {
	return
}

func (n *nilLogger) Errorln(v ...interface{}) {
	return
}

func (n *nilLogger) Fatal(v ...interface{}) {
	return
}

func (n *nilLogger) Fatalf(format string, v ...interface{}) {
	return
}

func (n *nilLogger) Fatalln(v ...interface{}) {
	return
}
