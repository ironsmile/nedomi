// This file is generated with go generate. Any changes to it will be lost after
// subsequent generates.
// If you want to edit it go to types.go.template and use `go generate` afterwards.

package upstream

import (
	"github.com/ironsmile/nedomi/config"

	"github.com/ironsmile/nedomi/upstream/simple"
)

type newUpstreamFunc func(*config.Config) Upstream

var upstreamTypes map[string]newUpstreamFunc = map[string]newUpstreamFunc{

	"simple": func(cfg *config.Config) Upstream {
		return simple.New(cfg)
	},
}
