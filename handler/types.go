// This file is generated with go generate. Any changes to it will be lost after
// subsequent generates.
// If you want to edit it go to types.go.template

package handler

import (
	"github.com/ironsmile/nedomi/config"

	"github.com/ironsmile/nedomi/handler/cache"
	"github.com/ironsmile/nedomi/handler/dir"
	"github.com/ironsmile/nedomi/handler/flv"
	"github.com/ironsmile/nedomi/handler/mp4"
	"github.com/ironsmile/nedomi/handler/purge"
	"github.com/ironsmile/nedomi/handler/status"
	"github.com/ironsmile/nedomi/handler/throttle"
	"github.com/ironsmile/nedomi/handler/via"
	"github.com/ironsmile/nedomi/types"
)

type newHandlerFunc func(*config.Handler, *types.Location, types.RequestHandler) (types.RequestHandler, error)

var handlerTypes = map[string]newHandlerFunc{

	"cache": func(cfg *config.Handler, l *types.Location, next types.RequestHandler) (types.RequestHandler, error) {
		return cache.New(cfg, l, next)
	},

	"dir": func(cfg *config.Handler, l *types.Location, next types.RequestHandler) (types.RequestHandler, error) {
		return dir.New(cfg, l, next)
	},

	"flv": func(cfg *config.Handler, l *types.Location, next types.RequestHandler) (types.RequestHandler, error) {
		return flv.New(cfg, l, next)
	},

	"mp4": func(cfg *config.Handler, l *types.Location, next types.RequestHandler) (types.RequestHandler, error) {
		return mp4.New(cfg, l, next)
	},

	"purge": func(cfg *config.Handler, l *types.Location, next types.RequestHandler) (types.RequestHandler, error) {
		return purge.New(cfg, l, next)
	},

	"status": func(cfg *config.Handler, l *types.Location, next types.RequestHandler) (types.RequestHandler, error) {
		return status.New(cfg, l, next)
	},

	"throttle": func(cfg *config.Handler, l *types.Location, next types.RequestHandler) (types.RequestHandler, error) {
		return throttle.New(cfg, l, next)
	},

	"via": func(cfg *config.Handler, l *types.Location, next types.RequestHandler) (types.RequestHandler, error) {
		return via.New(cfg, l, next)
	},
}
