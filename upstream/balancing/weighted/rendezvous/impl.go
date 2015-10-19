package rendezvous

import (
	"errors"
	"fmt"
	"hash/crc32"
	"sync"

	"github.com/ironsmile/nedomi/types"
)

type rbucket struct {
	types.UpstreamAddress
	weightPercent float64 // Values between 0 and 1
}
type rbuckets []rbucket

// Rendezvous implements upstream balancing that is based on the Rendezvous
// (a.k.a Highest Random Weight - HRW) hashing algorithm.
type Rendezvous struct {
	sync.RWMutex
	buckets rbuckets
}

// Set implements the balancing algorithm interface.
func (r *Rendezvous) Set(newBuckets []*types.UpstreamAddress) {
	r.Lock()
	defer r.Unlock()

	var totalWeight float64
	r.buckets = make(rbuckets, len(newBuckets))
	for i, b := range newBuckets {
		r.buckets[i] = rbucket{UpstreamAddress: *b}
		totalWeight += float64(b.Weight)
	}
	for i := range r.buckets {
		r.buckets[i].weightPercent = float64(r.buckets[i].Weight) / totalWeight
	}
}

// Get implements the balancing algorithm interface.
func (r *Rendezvous) Get(path string) (*types.UpstreamAddress, error) {
	r.RLock()
	defer r.RUnlock()
	if len(r.buckets) == 0 {
		return nil, errors.New("No upstream addresses set!")
	}

	found := false
	maxIdx := 0
	var maxScore float64

	//!TODO: implement O(log n) version: https://en.wikipedia.org/wiki/Rendezvous_hashing#Implementation
	for i := range r.buckets {
		key := []byte(r.buckets[i].ResolvedURL.Host + path)
		//!TODO: use faster and better-distributed algorithm than crc32? xxhash? murmur?
		score := r.buckets[i].weightPercent * float64(crc32.ChecksumIEEE(key))
		if score > maxScore {
			found = true
			maxIdx = i
			maxScore = score
		}
	}

	if !found {
		return nil, fmt.Errorf("No upstream addresses found for path %s", path)
	}

	return &r.buckets[maxIdx].UpstreamAddress, nil
}

// New creates a new Rendezvous upstream balancer.
func New() *Rendezvous {
	return &Rendezvous{}
}
