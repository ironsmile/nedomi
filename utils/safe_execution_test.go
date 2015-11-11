package utils

import (
	"strings"
	"testing"
	"time"
)

const panica = "паника"

func TestSafeExecute(t *testing.T) {
	t.Parallel()
	var ch = make(chan error)
	go SafeExecute(
		func() { panic(panica) },
		func(e error) { ch <- e },
	)

	err := <-ch
	if err == nil {
		t.Errorf("expected non nil error")
	}
	if !strings.HasPrefix(err.Error(), panica) {
		t.Errorf("expected the error to start with `%s` but it doesn't :\n%s",
			panica, err)
	}
}

func TestSafeExecuteWithNoPanic(t *testing.T) {
	t.Parallel()
	var ch = make(chan error)
	var ch2 = make(chan struct{})
	go SafeExecute(
		func() { ch2 <- struct{}{} },
		func(e error) { ch <- e },
	)

	select {
	case err := <-ch:
		t.Fatalf("got unexpected error: %s", err)
	case <-ch2:
	}

	select {
	case err := <-ch:
		t.Fatalf("got unexpected error: %s", err)
	case <-time.After(time.Millisecond * 10):
	}
}
