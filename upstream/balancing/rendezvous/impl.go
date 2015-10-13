package rendezvous

import "github.com/ironsmile/nedomi/upstream/balancing/random"

// New creates a new rendezvouz balancer.
func New() *random.Random {
	//!TODO: implement rendezvouz balancer
	return random.New()
}
