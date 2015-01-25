package app

import (
	"github.com/gophergala/nedomi/cache"
	"github.com/gophergala/nedomi/config"
)

/*
   Links a config vritual host to its cache manager
*/
type VirtualHost struct {
	config.VirtualHost
	CacheManger cache.CacheManager
}
