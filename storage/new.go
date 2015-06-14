/*
   The storageTypes map is in types.go and it is generate with `go generate`.
*/

//go:generate ./generate_storage_types

package storage

import (
	"fmt"

	"github.com/ironsmile/nedomi/cache"
	"github.com/ironsmile/nedomi/config"
	"github.com/ironsmile/nedomi/upstream"
)

/*
   New returns a new Storage ready for use. The st argument sets which type of storage
   will be returned.
*/
func New(st string, cfg config.CacheZoneSection, cm cache.CacheManager,
	up upstream.Upstream) (Storage, error) {

	storFunc, ok := storageTypes[st]

	if !ok {
		return nil, fmt.Errorf("No such storage type: %s", st)
	}

	return storFunc(cfg, cm, up), nil
}