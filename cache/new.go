// This file contains the function which returns a new Algorithm object
// based on its string name.
//
// New uses the cacheTypes map. This map is generated with
// `go generate` in the types.go file.

//go:generate go run ../tools/module_generator/main.go -template "types.go.template" -output "types.go"

package cache

import (
	"fmt"

	"github.com/ironsmile/nedomi/config"
)

// New creates and returns a particular type of cache algorithm.
func New(cz *config.CacheZoneSection) (Algorithm, error) {

	constructor, ok := cacheTypes[cz.Algorithm]
	if !ok {
		return nil, fmt.Errorf("No such cache algorithm: `%s` type", cz.Algorithm)
	}

	return constructor(cz), nil
}

// AlgorithmExists returns true if a Algorithm with this name exists.
// False otherwise.
func AlgorithmExists(ct string) bool {
	_, ok := cacheTypes[ct]
	return ok
}
