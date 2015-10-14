package roundrobin

import "github.com/ironsmile/nedomi/upstream/balancing/unweighted/random"

// New creates a new round-robin balancer.
func New() *random.Random {
	//!TODO: implement round-robin balancer
	return random.New()
}
