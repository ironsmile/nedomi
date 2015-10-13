// This file is generated with go generate. Any changes to it will be lost after
// subsequent generates.
// If you want to edit it go to types.go.template

package balancing

import (
	"github.com/ironsmile/nedomi/upstream/balancing/jump"
	"github.com/ironsmile/nedomi/upstream/balancing/ketama"
	"github.com/ironsmile/nedomi/upstream/balancing/random"
	"github.com/ironsmile/nedomi/upstream/balancing/rendezvous"
	"github.com/ironsmile/nedomi/upstream/balancing/roundrobin"
)

var balancingTypes = map[string]func() Algorithm{

	"jump": func() Algorithm {
		return jump.New()
	},

	"ketama": func() Algorithm {
		return ketama.New()
	},

	"random": func() Algorithm {
		return random.New()
	},

	"rendezvous": func() Algorithm {
		return rendezvous.New()
	},

	"roundrobin": func() Algorithm {
		return roundrobin.New()
	},
}
