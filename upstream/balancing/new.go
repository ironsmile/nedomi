package balancing

import "fmt"

// Package balancing deals algorithms that ballance HTTP requests between
// different upstream servers. It describes the Algorithm interface. Every
// subpackage *must* have a type which implements it.

// This file contains the function which returns a new Algorithm based on its ID.
//
// New uses the balancingTypes map. This map is generated with
// `go generate` in the types.go file.

//go:generate go run ../../tools/module_generator/main.go -template "types.go.template" -output "types.go"

// New creates and returns a new balancing algorithm
func New(t string) (Algorithm, error) {
	fnc, ok := balancingTypes[t]
	if !ok {
		return nil, fmt.Errorf("No such balancing algorithm module: %s", t)
	}

	return fnc(), nil
}
