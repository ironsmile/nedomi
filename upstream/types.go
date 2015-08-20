// This file is generated with go generate. Any changes to it will be lost after
// subsequent generates.
// If you want to edit it go to types.go.template and use `go generate` afterwards.

package upstream

import (
	"net/url"

	"github.com/ironsmile/nedomi/types"

	"github.com/ironsmile/nedomi/upstream/simple"
)

type newUpstreamFunc func(*url.URL) types.Upstream

var upstreamTypes = map[string]newUpstreamFunc{

	"simple": func(upstreamURL *url.URL) types.Upstream {
		return simple.New(upstreamURL)
	},
}
