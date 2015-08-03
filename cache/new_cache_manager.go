// This file contains the function which returns a new Manager object
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

// New creates and returns a particular type of cache manager.
func New(ct string, cz *config.CacheZoneSection) (Manager, error) {

	fnc, ok := cacheTypes[ct]

	if !ok {
		return nil, fmt.Errorf("No such cache manager: `%s` type", ct)
	}

	return fnc(cz), nil
}

// ManagerTypeExists returns true if a Manager with this name exists.
// False otherwise.
func ManagerTypeExists(ct string) bool {
	_, ok := cacheTypes[ct]
	return ok
}
