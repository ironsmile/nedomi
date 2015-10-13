package random

import (
	"math/rand"
	"sync"
	"time"

	"github.com/ironsmile/nedomi/types"
)

// Random randomly balances requests between its upstreams.
type Random struct {
	sync.RWMutex
	buckets []types.UpstreamAddress
	rnd     *rand.Rand
}

// Set implements the balancing algorithm interface.
func (r *Random) Set(buckets []types.UpstreamAddress) {
	r.Lock()
	defer r.Unlock()
	r.buckets = buckets
}

// Get implements the balancing algorithm interface.
func (r *Random) Get(_ string) types.UpstreamAddress {
	r.RLock()
	defer r.RUnlock()
	return r.buckets[r.rnd.Intn(len(r.buckets))]
}

// New creates a new random upstream balancer.
func New() *Random {
	return &Random{rnd: rand.New(rand.NewSource(time.Now().UnixNano()))}
}
