/*
   This file contains the function which returns a new RequestHandler
   based on its string name.

   New uses the handlerTypes map. This map is generated with
   `go generate` in the types.go file.
*/

//go:generate go run ../tools/module_generator/main.go -template "types.go.template" -output "types.go"

package handler

import (
	"fmt"
)

/*
   New creates and returns a new RequestHandler identified by its module name.
   Identifier is mhe module's directory (hence its package name).
*/
func New(module string) (RequestHandler, error) {

	fnc, ok := handlerTypes[module]

	if !ok {
		return nil, fmt.Errorf("No such request handler module: %s", module)
	}

	return fnc(), nil
}
