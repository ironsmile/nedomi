// The storageTypes map is in types.go and it is generate with `go generate`.

//go:generate go run ../tools/module_generator/main.go -template "types.go.template" -output "types.go"

package storage

import (
	"fmt"

	"github.com/ironsmile/nedomi/cache"
	"github.com/ironsmile/nedomi/config"
	"github.com/ironsmile/nedomi/logger"
	"github.com/ironsmile/nedomi/upstream"
)

// New returns a new Storage ready for use. The st argument sets which type of
// storage will be returned.
func New(st string, cfg config.CacheZoneSection, cm cache.Manager,
	up upstream.Upstream, logger logger.Logger) (Storage, error) {

	storFunc, ok := storageTypes[st]

	if !ok {
		return nil, fmt.Errorf("No such storage type: %s", st)
	}

	return storFunc(cfg, cm, up, logger), nil
}
