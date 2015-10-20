package roundrobin

import (
	"errors"
	"sync"
	"sync/atomic"

	"github.com/ironsmile/nedomi/types"
)

// RoundRobin balances requests between its upstreams one by one.
type RoundRobin struct {
	sync.RWMutex
	buckets []*types.UpstreamAddress
	counter uint32
}

// Set implements the balancing algorithm interface.
func (rr *RoundRobin) Set(buckets []*types.UpstreamAddress) {
	rr.Lock()
	defer rr.Unlock()
	rr.buckets = buckets
}

// Get implements the balancing algorithm interface.
func (rr *RoundRobin) Get(path string) (*types.UpstreamAddress, error) {
	rr.RLock()
	defer rr.RUnlock()
	if len(rr.buckets) == 0 {
		return nil, errors.New("No upstream addresses set!")
	}

	idx := atomic.AddUint32(&rr.counter, 1) % uint32(len(rr.buckets))
	return rr.buckets[idx], nil
}

// New creates a new round-robin upstream balancer.
func New() *RoundRobin {
	return &RoundRobin{}
}
