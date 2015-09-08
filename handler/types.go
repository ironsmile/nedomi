// This file is generated with go generate. Any changes to it will be lost after
// subsequent generates.
// If you want to edit it go to types.go.template

package handler

import (
	"encoding/json"

	"github.com/ironsmile/nedomi/handler/proxy"
	"github.com/ironsmile/nedomi/handler/status"
	"github.com/ironsmile/nedomi/types"
)

type newHandlerFunc func(*json.RawMessage, *types.Location) (types.RequestHandler, error)

var handlerTypes = map[string]newHandlerFunc{

	"proxy": func(cfg *json.RawMessage, l *types.Location) (types.RequestHandler, error) {
		return proxy.New(cfg, l)
	},

	"status": func(cfg *json.RawMessage, l *types.Location) (types.RequestHandler, error) {
		return status.New(cfg, l)
	},
}
