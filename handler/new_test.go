package handler

import "testing"

func TestErrorOnNonExistingHandler(t *testing.T) {
	_, err := New("bogus_handler", nil, nil)

	if err == nil {
		t.Error("No error returned with bogus handler.")
	}
}
