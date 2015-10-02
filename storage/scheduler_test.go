package storage

import (
	"testing"
	"time"

	"github.com/ironsmile/nedomi/types"
)

const DELTA = 10 * time.Millisecond

var fooKey = keyFromString("foo")

func writeFunc(ch chan string, str string) func() {
	return func() {
		ch <- str
	}
}

func keyFromString(s string) types.ObjectIDHash {
	var h = types.ObjectIDHash{}
	copy(h[:], s)
	return h
}

func TestAddingEvent(t *testing.T) {
	t.Parallel()
	defer func() {
		if v := recover(); v != nil {
			t.Errorf("Panic: %s", v)
		}
	}()
	mp := NewScheduler()
	defer mp.Destroy()

	expected := "bar"

	ch := make(chan string)
	mp.AddEvent(fooKey, writeFunc(ch, expected), time.Millisecond)
	found := <-ch

	if found != expected {
		t.Error("Found key was different than the expected")
	}
}

func TestMultipleSetsOnTheSameKey(t *testing.T) {
	t.Parallel()
	mp := NewScheduler()
	defer mp.Destroy()
	var bad, good = "bad", "good"

	ch := make(chan string)
	mp.AddEvent(fooKey, writeFunc(ch, bad), time.Second)
	mp.AddEvent(fooKey, writeFunc(ch, good), 100*time.Millisecond)

	got := waitAround(t, ch, 100*time.Millisecond)
	if !t.Failed() {
		return
	}
	if got != good {
		t.Errorf("expected '%s' got '%s'", good, got)
	}

	select {
	case got := <-ch:
		t.Errorf("expected nothing got '%s'", got)
	case <-time.After(1*time.Second + DELTA):
	}
}

func waitAround(t *testing.T, ch chan string, around time.Duration) string {
	var tooSoon = true
	for {
		select {
		case got := <-ch:
			if tooSoon {
				t.Fatal("return too fast")
				return ""
			}
			return got
		case <-time.After(around - DELTA):
			tooSoon = false
		case <-time.After(around + DELTA):
			t.Fatal("too slow")
			return ""
		}
	}
}

func TestContainsMethod(t *testing.T) {
	mp := NewScheduler()
	defer mp.Destroy()

	mp.AddEvent(fooKey, nil, 3*time.Second)

	if mp.Contains(fooKey) == false {
		t.Errorf("Contains: false negative with foo")
	}

	if mp.Contains(keyFromString("baba")) {
		t.Errorf("Contains: false positive with baba")
	}
}

func TestCleaningUpTheMap(t *testing.T) {
	mp := NewScheduler()
	defer mp.Destroy()

	expected := "pesho"
	ch := make(chan string)
	mp.AddEvent(fooKey, writeFunc(ch, expected), 10*time.Millisecond)

	mp.Cleanup()
	select {
	case got := <-ch:
		t.Errorf("expected nothing got '%s'", got)
	case <-time.After(1*time.Second + DELTA):
	}
}

func TestPanics(t *testing.T) {
	t.Parallel()
	mp := NewScheduler()
	defer mp.Destroy()

	mp.AddEvent(fooKey, func() {
		panic("bar")
	}, time.Millisecond)

	<-time.After(1*time.Millisecond + DELTA)
	if mp.Contains(fooKey) {
		t.Error("the panicing function has not expired")
	}
}
