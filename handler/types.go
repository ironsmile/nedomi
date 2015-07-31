// This file is generated with go generate. Any changes to it will be lost after
// subsequent generates.
// If you want to edit it go to types.go.template

package handler

import (
	"github.com/ironsmile/nedomi/handler/proxy"
	"github.com/ironsmile/nedomi/handler/status"
)

type newHandlerFunc func() RequestHandler

var handlerTypes = map[string]newHandlerFunc{

	"proxy": func() RequestHandler {
		return proxy.New()
	},

	"status": func() RequestHandler {
		return status.New()
	},
}
