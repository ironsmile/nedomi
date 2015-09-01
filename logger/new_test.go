package logger

import (
	"testing"

	"github.com/ironsmile/nedomi/config"
)

func TestCreatingBogusLogger(t *testing.T) {
	_, err := New(&config.LoggerSection{Type: "bogus_logger"})

	if err == nil {
		t.Error("There was no error when creating bogus logger")
	}
}
