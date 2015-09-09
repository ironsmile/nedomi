package handler

import (
	"encoding/json"
	"testing"
)

func TestErrorOnNonExistingHandler(t *testing.T) {
	_, err := New("bogus_handler", json.RawMessage{}, nil)

	if err == nil {
		t.Error("No error returned with bogus handler.")
	}
}
