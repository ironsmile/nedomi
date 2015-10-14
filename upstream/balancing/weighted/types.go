// This file is generated with go generate. Any changes to it will be lost after
// subsequent generates.
// If you want to edit it go to types.go.template

package weighted

import (
	"github.com/ironsmile/nedomi/types"
	"github.com/ironsmile/nedomi/upstream/balancing/weighted/ketama"
	"github.com/ironsmile/nedomi/upstream/balancing/weighted/random"
	"github.com/ironsmile/nedomi/upstream/balancing/weighted/rendezvous"
)

// Algorithms contains all weighted upstream balancing algorithm implementations.
var Algorithms = map[string]func() types.UpstreamBalancingAlgorithm{

	"ketama": func() types.UpstreamBalancingAlgorithm {
		return ketama.New()
	},

	"random": func() types.UpstreamBalancingAlgorithm {
		return random.New()
	},

	"rendezvous": func() types.UpstreamBalancingAlgorithm {
		return rendezvous.New()
	},
}
