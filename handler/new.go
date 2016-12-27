// Package handler deals with the main HTTP handler modules for nedomi. It describes the
// RequestHandler interface. Every subpackage *must* have a type which implements it.

// This file contains the function which returns a new RequestHandler
// based on its string name.
//
// New uses the handlerTypes map. This map is generated with
// `go generate` in the types.go file.

//go:generate go run ../tools/module_generator/main.go -template "types.go.template" -output "types.go"
//go:generate go run ../tools/module_generator/main.go -inputlist additional.list -template "additional_types.go.template" -output "additional_types.go"

package handler

import (
	"fmt"

	"github.com/ironsmile/nedomi/config"
	"github.com/ironsmile/nedomi/types"
)

type newHandlerFunc func(*config.Handler, *types.Location, types.RequestHandler) (types.RequestHandler, error)

// New creates and returns a new RequestHandler identified by its module name.
// Identifier is the module's directory (hence its package name).
// Additionally it receives handler specific config in the form of *json.RawMessage
// and types.Location representing the location the handler will be used for.
func New(cfg *config.Handler, l *types.Location, next types.RequestHandler) (types.RequestHandler, error) {
	if cfg == nil {
		return nil, fmt.Errorf("handler.New requires a non nil configuration")
	}
	fnc, ok := handlerTypes[cfg.Type]
	if !ok {
		return nil, fmt.Errorf("no such request handler module: %s", cfg.Type)
	}

	return fnc(cfg, l, next)
}
