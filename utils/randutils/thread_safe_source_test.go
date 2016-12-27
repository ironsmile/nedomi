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

	var source = NewThreadSafeSource()
	rnd := rand.New(source)
	for i := 0; i < 100; i++ {
		go func() {
			for j := 0; j < 200; j++ {
				rnd.Int()
			}
			wg.Done()
		}()
		go func(i int64) {
			source.Seed(i)
		}(int64(i))
	}

	wg.Wait()
}
