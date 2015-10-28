// This file is generated with go generate. Any changes to it will be lost after
// subsequent generates.
// If you want to edit it go to types.go.template

package unweighted

import (
	"github.com/ironsmile/nedomi/types"
	"github.com/ironsmile/nedomi/upstream/balancing/unweighted/random"
	"github.com/ironsmile/nedomi/upstream/balancing/unweighted/roundrobin"
)

// Algorithms contains all unweighted upstream balancing algorithm implementations.
var Algorithms = map[string]func() types.UpstreamBalancingAlgorithm{

	"random": func() types.UpstreamBalancingAlgorithm {
		return random.New()
	},

	"roundrobin": func() types.UpstreamBalancingAlgorithm {
		return roundrobin.New()
	},
}
