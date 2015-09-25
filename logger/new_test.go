package logger

import (
	"testing"

	"github.com/ironsmile/nedomi/config"
)

func TestCreatingLoggerWithNilCfg(t *testing.T) {
	t.Parallel()
	_, err := New(nil)

	if err == nil {
		t.Error("There was no error when creating logger with a nil config")
	}
}

func TestCreatingBogusLogger(t *testing.T) {
	t.Parallel()
	_, err := New(config.NewLogger("bogus_logger", nil))

	if err == nil {
		t.Error("There was no error when creating bogus logger")
	}
}
