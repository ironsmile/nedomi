package random

import (
	"errors"
	"math/rand"
	"sort"
	"sync"

	"github.com/ironsmile/nedomi/types"
	"github.com/ironsmile/nedomi/utils/randutils"
)

type rbuckets []*types.UpstreamAddress

// Implement sort.Interface for rbuckets
func (rb rbuckets) Len() int           { return len(rb) }
func (rb rbuckets) Less(i, j int) bool { return rb[i].Weight > rb[j].Weight }
func (rb rbuckets) Swap(i, j int)      { rb[i], rb[j] = rb[j], rb[i] }

// Random randomly balances requests between its upstreams.
type Random struct {
	sync.RWMutex
	buckets     rbuckets
	totalWeight int
	rnd         *rand.Rand
}

// Set implements the balancing algorithm interface.
func (r *Random) Set(buckets []*types.UpstreamAddress) {
	r.Lock()
	defer r.Unlock()
	r.buckets = buckets
	sort.Sort(r.buckets)

	r.totalWeight = 0
	for _, b := range r.buckets {
		r.totalWeight += int(b.Weight)
	}
}

// Get implements the balancing algorithm interface.
func (r *Random) Get(_ string) (*types.UpstreamAddress, error) {
	r.RLock()
	defer r.RUnlock()

	if r.totalWeight <= 0 {
		return nil, errors.New("No configured upstreams or upstream weights")
	}

	chosen := uint32(r.rnd.Intn(r.totalWeight))
	var reachedWeight uint32

	for _, b := range r.buckets {
		reachedWeight += b.Weight
		if chosen < reachedWeight {
			return b, nil
		}
	}

	return nil, errors.New("Could not get a weighted random upstream address")
}

// New creates a new weighted random upstream balancer.
func New() *Random {
	return &Random{rnd: rand.New(randutils.NewThreadSafeSource())}
}
