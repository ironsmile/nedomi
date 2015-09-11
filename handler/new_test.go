package handler

import (
	"testing"

	"github.com/ironsmile/nedomi/config"
)

func TestErrorOnNilHandlerConfig(t *testing.T) {
	_, err := New(nil, nil, nil)

	if err == nil {
		t.Error("No error returned with bogus handler.")
	}
}

func TestErrorOnNonExistingHandler(t *testing.T) {
	_, err := New(config.NewHandler("bogus_handler", nil), nil, nil)

	if err == nil {
		t.Error("No error returned with bogus handler.")
	}
}
