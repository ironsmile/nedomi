package jump

import "github.com/ironsmile/nedomi/upstream/balancing/random"

// New creates a new Jump balancer
func New() *random.Random {
	//!TODO: implement https://github.com/dgryski/go-jump
	return random.New()
}
