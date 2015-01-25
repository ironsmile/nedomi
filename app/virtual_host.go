package app

import (
	"github.com/gophergala/nedomi/cache"
	"github.com/gophergala/nedomi/config"
	"github.com/gophergala/nedomi/storage"
)

/*
   Links a config vritual host to its cache manager
*/
type VirtualHost struct {
	config.VirtualHost
	CacheManger cache.CacheManager
	Storage     storage.Storage
}
