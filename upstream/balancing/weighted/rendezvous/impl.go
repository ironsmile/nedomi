package rendezvous

import (
	"errors"
	"fmt"
	"hash/crc32"
	"math"
	"sync"

	"github.com/ironsmile/nedomi/types"
)

type rbucket struct {
	types.UpstreamAddress
	weightPercent    float64 // Values between 0 and 1
	weightMultiplier float64
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

	var PLast, Xn, XLast, Kk1, weight, x float64
	Xn = 1.0
	K := len(r.buckets)
	//!TODO: understand how this works :)
	for k := 1; k <= K; k++ {
		Kk1 = float64(K - k + 1)
		weight = r.buckets[k-1].weightPercent
		x = Kk1 * (weight - PLast) / Xn
		x += math.Pow(XLast, Kk1)
		r.buckets[k-1].weightMultiplier = math.Pow(x, 1.0/Kk1)
		Xn *= r.buckets[k-1].weightMultiplier
		XLast = r.buckets[k-1].weightMultiplier
		PLast = weight
	}
}

// Get implements the balancing algorithm interface.
func (r *Rendezvous) Get(path string) (*types.UpstreamAddress, error) {
	r.RLock()
	defer r.RUnlock()
	if len(r.buckets) == 0 {
		return nil, errors.New("no upstream addresses set")
	}

	found := false
	maxIdx := 0
	var maxScore float64

	for i := range r.buckets {
		key := []byte(r.buckets[i].Host + path)
		//!TODO: use faster and better-distributed algorithm than crc32? xxhash? murmur?
		score := r.buckets[i].weightMultiplier * float64(crc32.ChecksumIEEE(key))
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
