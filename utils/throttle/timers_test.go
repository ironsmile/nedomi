package throttle

import (
	"sync"
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

func testSleepWithPooledTimer(t testing.TB, sleepFor func(d time.Duration) <-chan struct{}) {
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

func TestSleepWithPooledTimer(t *testing.T) {
	testSleepWithPooledTimer(t, stdSleepFor)
}

func TestParallelSleepWithPooledTimer(t *testing.T) {
	var wg = make(chan struct{}, parallelSleeps)

	for i := 0; parallelSleeps > i; i++ {
		go func(i int) {
			testSleepWithPooledTimer(t, pooledSleepFor)
			wg <- struct{}{}
		}(i)
	}

	for i := 0; parallelSleeps > i && !t.Failed(); i++ {
		<-wg
	}
}

func stdSleepFor(d time.Duration) <-chan struct{} {
	var ch = make(chan struct{})
	go func() {
		time.Sleep(d)
		sendStructForASecond(ch, d*2)
	}()
	return ch
}

func pooledSleepFor(d time.Duration) <-chan struct{} {
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

func benchmarkParallelSleepTemplate(b *testing.B, sleepFor func(d time.Duration)) {
	var wg sync.WaitGroup
	for j := 0; j < b.N; j++ {
		wg.Add(10000 * 2)
		for i := 0; 10000 > i; i++ {
			go func() {
				sleepFor(time.Millisecond * 5)
				wg.Done()
			}()

			go func() {
				sleepFor(time.Millisecond * 50)
				wg.Done()
			}()
		}
		wg.Wait()
	}
}

func BenchmarkParallelSleepWithSTD(b *testing.B) {
	benchmarkParallelSleepTemplate(b, time.Sleep)
}

func BenchmarkParallelSleepWithPooledTimer(b *testing.B) {
	benchmarkParallelSleepTemplate(b, sleepWithPooledTimer)
}
