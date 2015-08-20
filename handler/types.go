// This file is generated with go generate. Any changes to it will be lost after
// subsequent generates.
// If you want to edit it go to types.go.template

package handler

import (
	"github.com/ironsmile/nedomi/handler/proxy"
	"github.com/ironsmile/nedomi/handler/status"
	"github.com/ironsmile/nedomi/types"
)

type newHandlerFunc func() types.RequestHandler

var handlerTypes = map[string]newHandlerFunc{

	"proxy": func() types.RequestHandler {
		return proxy.New()
	},

	"status": func() types.RequestHandler {
		return status.New()
	},
}
