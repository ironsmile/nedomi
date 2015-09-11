// Package cache is implements the caching algorithm. It defines the Algorithm
// interface. Every CacheZone has its own cache algorithm. This makes it possible for
// different caching algorithms to be used in the same time.

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
	"github.com/ironsmile/nedomi/types"
)

// New creates and returns a particular type of cache algorithm.
func New(cz *config.CacheZone, remove func(*types.ObjectIndex) error,
	logger types.Logger) (types.CacheAlgorithm, error) {

	if cz == nil {
		return nil, fmt.Errorf("Empty cache zone configuration supplied!")
	}

	//!TODO: do not pass the whole CZ config object, just the algorithm details
	constructor, ok := cacheTypes[cz.Algorithm]
	if !ok {
		return nil, fmt.Errorf("No such cache algorithm: `%s` type", cz.Algorithm)
	}

	return constructor(cz, remove, logger), nil
}
