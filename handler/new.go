// Package handler deals with the main HTTP handler modules for nedomi. It describes the
// RequestHandler interface. Every subpackage *must* have a type which implements it.

// This file contains the function which returns a new RequestHandler
// based on its string name.
//
// New uses the handlerTypes map. This map is generated with
// `go generate` in the types.go file.

//go:generate go run ../tools/module_generator/main.go -template "types.go.template" -output "types.go"

package handler

import (
	"encoding/json"
	"fmt"

	"github.com/ironsmile/nedomi/types"
)

// New creates and returns a new RequestHandler identified by its module name.
// Identifier is the module's directory (hence its package name).
// Additionaly it receives handler specific config in the form of *json.RawMessage
// and types.Location representing the location the handler will be used for.
func New(module string, cfg *json.RawMessage, l *types.Location) (types.RequestHandler, error) {
	fnc, ok := handlerTypes[module]
	if !ok {
		return nil, fmt.Errorf("No such request handler module: %s", module)
	}

	return fnc(cfg, l)
}
