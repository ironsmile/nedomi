package random

import (
	"errors"
	"math/rand"
	"sync"

	"github.com/ironsmile/nedomi/types"
	"github.com/ironsmile/nedomi/utils/randutils"
)

// Random randomly balances requests between its upstreams.
type Random struct {
	sync.RWMutex
	buckets []*types.UpstreamAddress
	rnd     *rand.Rand
}

// Set implements the balancing algorithm interface.
func (r *Random) Set(buckets []*types.UpstreamAddress) {
	r.Lock()
	defer r.Unlock()
	r.buckets = buckets
}

// Get implements the balancing algorithm interface.
func (r *Random) Get(_ string) (*types.UpstreamAddress, error) {
	r.RLock()
	defer r.RUnlock()
	if len(r.buckets) == 0 {
		return nil, errors.New("No upstream addresses set!")
	}

	return r.buckets[r.rnd.Intn(len(r.buckets))], nil
}

// New creates a new random upstream balancer.
func New() *Random {
	return &Random{rnd: rand.New(randutils.NewThreadSafeSource())}
}
