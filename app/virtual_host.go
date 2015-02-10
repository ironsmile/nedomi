package app

import (
	"github.com/ironsmile/nedomi/cache"
	"github.com/ironsmile/nedomi/config"
	"github.com/ironsmile/nedomi/storage"
)

/*
   Links a config vritual host to its cache manager and a storage object.
*/
type VirtualHost struct {
	config.VirtualHost
	CacheManger cache.CacheManager
	Storage     storage.Storage
}
