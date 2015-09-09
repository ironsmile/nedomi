package logger

import (
	"testing"

	"github.com/ironsmile/nedomi/config"
)

func TestCreatingLoggerWithNilCfg(t *testing.T) {
	_, err := New(nil)

	if err == nil {
		t.Error("There was no error when creating logger with a nil config")
	}
}

func TestCreatingBogusLogger(t *testing.T) {
	_, err := New(&config.Logger{Type: "bogus_logger"})

	if err == nil {
		t.Error("There was no error when creating bogus logger")
	}
}
