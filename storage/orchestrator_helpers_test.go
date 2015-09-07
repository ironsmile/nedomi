package storage

import (
	"testing"
	"time"

	"github.com/ironsmile/nedomi/cache"
	"github.com/ironsmile/nedomi/logger"
)

func wait(t *testing.T, period time.Duration, errorMessage string, action func()) {
	finished := make(chan struct{})

	go func() {
		action()
		close(finished)
	}()

	select {
	case <-finished:
	case <-time.After(period):
		t.Errorf("Test exceeded allowed time of %d seconds: %s", period/time.Second, errorMessage)
	}
	return
}

func getTestOrchestrator() (*Orchestrator, chan<- struct{}) {
	doneCh := make(chan struct{})
	o := &Orchestrator{
		storage:      NewMock(),
		algorithm:    cache.NewMock(nil),
		logger:       logger.NewMock(),
		foundObjects: make(chan *storageItem),
		done:         doneCh,
	}

	return o, doneCh
}

func TestConcurrentIteratorWithEmpty(t *testing.T) {
	t.Parallel()
	t.SkipNow()

	wait(t, 2*time.Second, "Iterating should not have hanged", func() {
		o, _ := getTestOrchestrator()
		o.startConcurrentIterator()
		if res, ok := <-o.foundObjects; ok {
			t.Errorf("The iterator did not close the channel immediately with an empty storage: %#v", res)
		}
	})
}

func TestConcurrentIteratorWithItems(t *testing.T) {
	t.Parallel()
	t.SkipNow()

	wait(t, 2*time.Second, "Iterating should not have hanged", func() {
		o, _ := getTestOrchestrator()
		o.storage.SaveMetadata(obj1)
		o.storage.SaveMetadata(obj2)
		o.startConcurrentIterator()

		if res1, ok := <-o.foundObjects; !ok || (res1.Obj != obj1 && res1.Obj != obj2) {
			t.Errorf("Iterator did not return object correctly: %#v", res1)
		}

		if res2, ok := <-o.foundObjects; !ok || (res2.Obj != obj1 && res2.Obj != obj2) {
			t.Errorf("Iterator did not return object correctly: %#v", res2)
		}

		if res, ok := <-o.foundObjects; ok {
			t.Errorf("The iterator did not close the channel after all elements were returned: %#v", res)
		}
	})
}

func TestConcurrentIteratorCancel(t *testing.T) {
	t.Parallel()
	t.SkipNow()

	wait(t, 2*time.Second, "Iterating should not have hanged", func() {
		o, doneCh := getTestOrchestrator()
		o.storage.SaveMetadata(obj1)
		o.storage.SaveMetadata(obj2)
		o.startConcurrentIterator()

		if res1, ok := <-o.foundObjects; !ok || (res1.Obj != obj1 && res1.Obj != obj2) {
			t.Errorf("Iterator did not return object correctly: %#v", res1)
		}
		close(doneCh)
		time.Sleep(200 * time.Millisecond)

		if res, ok := <-o.foundObjects; ok {
			t.Errorf("The iterator did not close the channel after the cancel: %#v", res)
		}
	})
}
