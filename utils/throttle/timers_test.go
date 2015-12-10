package throttle

import (
	"testing"
	"time"
)

const (
	sleepTime         = time.Millisecond * 50
	fastWait          = time.Millisecond * 30
	slowWait          = time.Millisecond * 30
	parallelSleeps    = 20
	sleepsInGoroutine = 100
)

func TestSleepWithPooledTimer(t *testing.T) {
	for i := 0; sleepsInGoroutine > i; i++ {
		var ch = sleepFor(sleepTime)
		select {
		case <-ch:
			t.Fatalf("sleep was too fast %d", i)
		case <-time.After(fastWait):
			// all is fine in this world
		}
		select {
		case <-ch:
			// all is fine in this world
		case <-time.After(slowWait):
			t.Fatalf("sleep was too slow %d", i)
		}
	}
}

func TestParallelSleepWithPooledTimer(t *testing.T) {
	var wg = make(chan struct{}, parallelSleeps)

	for i := 0; parallelSleeps > i; i++ {
		go func(i int) {
			TestSleepWithPooledTimer(t)
			wg <- struct{}{}
		}(i)
	}

	for i := 0; parallelSleeps > i && !t.Failed(); i++ {
		<-wg
	}
}

func sleepFor(d time.Duration) <-chan struct{} {
	var ch = make(chan struct{})
	go func() {
		sleepWithPooledTimer(d)
		sendStructForASecond(ch, d*2)
	}()
	return ch
}

func sendStructForASecond(ch chan struct{}, d time.Duration) {
	select {
	case ch <- struct{}{}:
	case <-time.After(d):
	}
}
