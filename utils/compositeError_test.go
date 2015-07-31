package utils

import (
	"errors"
	"testing"
)

func TestCompositeErrorString(t *testing.T) {
	err := &CompositeError{}

	if !err.Empty() {
		t.Error("Composite Error was not empty after created")
	}

	err.AppendError(errors.New("First Error"))
	err.AppendError(errors.New("Seconds Error"))

	if err.Empty() {
		t.Error("Composite error was empty after adding two errors")
	}

	found := err.Error()

	expected := "First Error\nSeconds Error"

	if found != expected {
		t.Errorf("Expected error `%s` but it was `%s`", expected, found)
	}
}
