package randutils

import (
	"math/rand"
	"sync"
	"time"
)

// NewThreadSafeSource creates and returnes a new thread safe random source. It
// is seeded by the current unix timestamp.
func NewThreadSafeSource() rand.Source {
	return &lockedSource{src: rand.NewSource(time.Now().UnixNano())}
}

type lockedSource struct {
	sync.Mutex
	src rand.Source
}

func (r *lockedSource) Int63() (n int64) {
	r.Lock()
	n = r.src.Int63()
	r.Unlock()
	return
}

func (r *lockedSource) Seed(seed int64) {
	r.Lock()
	r.src.Seed(seed)
	r.Unlock()
}
