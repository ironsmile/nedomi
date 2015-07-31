package logger

import (
	"testing"

	"github.com/ironsmile/nedomi/config"
)

func TestCreatingBogusLogger(t *testing.T) {
	_, err := New("bogus_logger", config.LoggerSection{})

	if err == nil {
		t.Error("There was no error when creating bogus logger")
	}
}
