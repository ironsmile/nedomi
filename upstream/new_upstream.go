// This file contains the function which returns a new Upstream object
// based on its string name.
//
// New uses the upstreamTypes map. This map is generated with
// `go generate` in the types.go file.

//go:generate go run ../tools/module_generator/main.go -template "types.go.template" -output "types.go"

package upstream

import (
	"fmt"
	"net/url"
)

// New creates and returns a particular type of upstream.
func New(upstreamName string, upstreamURL *url.URL) (Upstream, error) {

	fnc, ok := upstreamTypes[upstreamName]

	if !ok {
		return nil, fmt.Errorf("No such upstream type: `%s`", upstreamName)
	}

	return fnc(upstreamURL), nil
}

// TypeExists returns true if a upstream module with this name exists. False otherwise.
func TypeExists(upstreamName string) bool {
	_, ok := upstreamTypes[upstreamName]
	return ok
}
