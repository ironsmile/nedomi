// This file is generated with go generate. Any changes to it will be lost after
// subsequent generates.
// If you want to edit it go to types.go.template

package handler

import (
	"github.com/ironsmile/nedomi/config"

	"github.com/ironsmile/nedomi/handler/proxy"
	"github.com/ironsmile/nedomi/handler/status"
	"github.com/ironsmile/nedomi/handler/via"
	"github.com/ironsmile/nedomi/types"
)

type newHandlerFunc func(*config.Handler, *types.Location, types.RequestHandler) (types.RequestHandler, error)

var handlerTypes = map[string]newHandlerFunc{

	"proxy": func(cfg *config.Handler, l *types.Location, next types.RequestHandler) (types.RequestHandler, error) {
		return proxy.New(cfg, l, next)
	},

	"status": func(cfg *config.Handler, l *types.Location, next types.RequestHandler) (types.RequestHandler, error) {
		return status.New(cfg, l, next)
	},

	"via": func(cfg *config.Handler, l *types.Location, next types.RequestHandler) (types.RequestHandler, error) {
		return via.New(cfg, l, next)
	},
}
