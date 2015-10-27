package jump

import (
	"errors"
	"sync"

	"github.com/ironsmile/nedomi/types"

	xxhash "github.com/OneOfOne/xxhash/native"
)

// Jump implements the jump consistent hash algorithm from the paper "A Fast,
// Minimal Memory, Consistent Hash Algorithm" by John Lamping, Eric Veach (2014)
// http://arxiv.org/abs/1406.2294
type Jump struct {
	sync.RWMutex
	buckets []*types.UpstreamAddress
}

// Set implements the balancing algorithm interface.
func (j *Jump) Set(buckets []*types.UpstreamAddress) {
	j.Lock()
	defer j.Unlock()
	j.buckets = buckets
}

// Get implements the balancing algorithm interface.
func (j *Jump) Get(path string) (*types.UpstreamAddress, error) {
	j.RLock()
	defer j.RUnlock()

	len := int64(len(j.buckets))
	if len == 0 {
		return nil, errors.New("No upstream addresses set!")
	}

	key := xxhash.Checksum64([]byte(path))

	var b int64 = -1
	var i int64

	for i < len {
		b = i
		key = key*2862933555777941757 + 1
		i = int64(float64(b+1) * (float64(int64(1)<<31) / float64((key>>33)+1)))
	}

	// The above algorithm sets i in the inclusive interval [0, len(buckets)]
	return j.buckets[i%len], nil
}

// New creates a new Jump balancer.
func New() *Jump {
	return &Jump{}
}
