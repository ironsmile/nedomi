/*
   This file contains the function which returns a new Upstream object
   based on its string name.

   New uses the upstreamTypes map. This map is generated with
   `go generate` in the types.go file.
*/

//go:generate go run ../tools/module_generator/main.go -template "types.go.template" -output "types.go"

package upstream

import (
	"fmt"

	"github.com/ironsmile/nedomi/config"
)

/*
   New creates and returns a particular type of cache manager.
*/
func New(upstreamName string, cfg *config.Config) (Upstream, error) {

	fnc, ok := upstreamTypes[upstreamName]

	if !ok {
		return nil, fmt.Errorf("No such upstream type: `%s`", upstreamName)
	}

	return fnc(cfg), nil
}

/*
   Returns true if a upstream module with this name exists. False otherwise.
*/
func UpstreamTypeExists(upstreamName string) bool {
	_, ok := upstreamTypes[upstreamName]
	return ok
}
