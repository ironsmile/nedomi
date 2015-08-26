package utils

import (
	"errors"
	"testing"
)

func TestConstructor(t *testing.T) {
	err := NewCompositeError(errors.New("1"), errors.New("2"), nil)
	if len(*err) != 2 || err.Error() != "1\n2" {
		t.Error("Composite error constructor does not work")
	}
}

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

func TestCompositeErrorWithNil(t *testing.T) {
	err := &CompositeError{}

	err.AppendError(nil)
	if !err.Empty() {
		t.Error("Composite Error was not empty after adding a nil error")
	}

	err.AppendError(errors.New("First Error"))

	if err.Empty() {
		t.Error("Composite error was empty after adding an error")
	}

	err.AppendError(nil)

	found := err.Error()

	expected := "First Error"

	if found != expected {
		t.Errorf("Expected error `%s` but it was `%s`", expected, found)
	}
}
