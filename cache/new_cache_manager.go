// This file contains the function which returns a new CacheManager object
// based on its string name.
//
// NewCacheManager uses the cacheTypes map. This map is generated with
// `go generate` in the types.go file.

//go:generate go run ../tools/module_generator/main.go -template "types.go.template" -output "types.go"

package cache

import (
	"fmt"

	"github.com/ironsmile/nedomi/config"
)

// NewCacheManager creates and returns a particular type of cache manager.
func NewCacheManager(ct string, cz *config.CacheZoneSection) (CacheManager, error) {

	fnc, ok := cacheTypes[ct]

	if !ok {
		return nil, fmt.Errorf("No such cache manager: `%s` type", ct)
	}

	return fnc(cz), nil
}

// CacheManagerTypeExists returns true if a CacheManager with this name exists.
// False otherwise.
func CacheManagerTypeExists(ct string) bool {
	_, ok := cacheTypes[ct]
	return ok
}
