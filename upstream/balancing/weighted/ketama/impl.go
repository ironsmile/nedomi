package ketama

import (
	"crypto/md5"
	"fmt"
	"sort"
	"sync"

	"github.com/ironsmile/nedomi/types"
)

const (
	pointsPerServer = 160
	pointsPerHash   = 4
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

var getPointHash = md5.Sum

func getKetamaHash(key string, alignment uint32) uint32 {
	digest := getPointHash([]byte(key))

	return uint32(digest[3+alignment*4])<<24 |
		uint32(digest[2+alignment*4])<<16 |
		uint32(digest[1+alignment*4])<<8 |
		uint32(digest[alignment*4])
}

// Set implements the balancing algorithm interface.
func (k *Ketama) Set(upstreams []*types.UpstreamAddress) {
	k.Lock()
	defer k.Unlock()

	upstreamsCount := uint32(len(upstreams))
	maxPoints := upstreamsCount * pointsPerServer
	k.ring = make([]continuumPoint, 0, maxPoints)

	var totalWeight uint32
	for _, u := range upstreams {
		totalWeight += u.Weight
	}

	for _, u := range upstreams {
		points := (u.Weight * maxPoints) / totalWeight
		if points%pointsPerHash != 0 {
			points += pointsPerHash - points%pointsPerHash
		}

		preSource := u.Hostname
		if u.Port != "80" {
			preSource = u.Hostname + ":" + u.Port
		}

		for i := uint32(0); i < points/pointsPerHash; i++ {
			source := fmt.Sprintf("%s-%dd", preSource, i)

			for x := uint32(0); x < pointsPerHash; x++ {
				if uint32(len(k.ring)) >= maxPoints {
					continue
				}

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
