package ketama

import (
	"crypto/md5"
	"errors"
	"fmt"
	"sort"
	"sync"

	"github.com/ironsmile/nedomi/types"
)

const (
	pointsPerServer = 160
	pointsPerHash   = md5.Size / 4
)

type continuumPoint struct {
	*types.UpstreamAddress
	point uint32
}

type continuumPoints []continuumPoint

// Implement sort.Interface for rbuckets
func (cp continuumPoints) Len() int           { return len(cp) }
func (cp continuumPoints) Less(i, j int) bool { return cp[i].point > cp[j].point }
func (cp continuumPoints) Swap(i, j int)      { cp[i], cp[j] = cp[j], cp[i] }

// Ketama implements upstream balancing that is based on the Ketama consistent
// hashing algorithm.
type Ketama struct {
	sync.RWMutex
	ring continuumPoints
}

func getKetamaHash(key string, alignment uint32) uint32 {
	digest := md5.Sum([]byte(key))

	return uint32(digest[3+alignment*4])<<24 |
		uint32(digest[2+alignment*4])<<16 |
		uint32(digest[1+alignment*4])<<8 |
		uint32(digest[alignment*4])
}

// Set implements the balancing algorithm interface.
func (k *Ketama) Set(upstreams []*types.UpstreamAddress) {
	k.Lock()
	defer k.Unlock()

	maxPoints := uint32(len(upstreams)) * pointsPerServer
	k.ring = make([]continuumPoint, 0, maxPoints)

	var totalWeight uint32
	for _, u := range upstreams {
		totalWeight += u.Weight
	}

	for _, u := range upstreams {
		points := (u.Weight * maxPoints) / totalWeight

		for i := uint32(0); i < points/pointsPerHash; i++ {
			source := fmt.Sprintf("%s:%s-%d", u.Hostname, u.Port, i)

			for x := uint32(0); x < pointsPerHash; x++ {

				k.ring = append(k.ring, continuumPoint{
					UpstreamAddress: u,
					point:           getKetamaHash(source, x),
				})
			}
		}
	}
	sort.Sort(k.ring)
}

// Get implements the balancing algorithm interface.
func (k *Ketama) Get(path string) (*types.UpstreamAddress, error) {
	k.RLock()
	defer k.RUnlock()

	if len(k.ring) == 0 {
		return nil, errors.New("No configured upstreams or upstream weights")
	}

	point := getKetamaHash(path, 0)
	seeker := func(i int) bool {
		return k.ring[i].point < point
	}

	idx := sort.Search(len(k.ring), seeker)
	if idx < len(k.ring) {
		return k.ring[idx].UpstreamAddress, nil
	}

	return k.ring[0].UpstreamAddress, nil
}

// New creates a new ketama consistent hash upstream balancer.
func New() *Ketama {
	return &Ketama{}
}
