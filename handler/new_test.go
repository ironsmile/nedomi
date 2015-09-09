package handler

import (
	"testing"

	"github.com/ironsmile/nedomi/config"
)

func TestErrorOnNilHandlerConfig(t *testing.T) {
	_, err := New(nil, nil)

	if err == nil {
		t.Error("No error returned with bogus handler.")
	}
}
func TestErrorOnNonExistingHandler(t *testing.T) {
	_, err := New(&config.Handler{Type: "bogus_handler"}, nil)

	if err == nil {
		t.Error("No error returned with bogus handler.")
	}
}
