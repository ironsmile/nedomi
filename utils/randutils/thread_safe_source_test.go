package randutils

import (
	"math/rand"
	"sync"
	"testing"
)

// This test will probably only be useful if `go test -race` is used
func TestRandomConcurrentUsage(t *testing.T) {
	t.Parallel()
	wg := sync.WaitGroup{}
	wg.Add(100)

	rnd := rand.New(NewThreadSafeSource())
	for i := 0; i < 100; i++ {
		go func() {
			rnd.Int()
			wg.Done()
		}()
	}

	wg.Wait()
}
