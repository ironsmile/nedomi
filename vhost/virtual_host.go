// Package vhost exports the VirtualHost type. This type represents a virtual
// host in the config.
package vhost

import (
	"github.com/ironsmile/nedomi/cache"
	"github.com/ironsmile/nedomi/config"
	"github.com/ironsmile/nedomi/storage"
)

// VirtualHost links a config vritual host to its cache algorithm and a storage object.
type VirtualHost struct {
	config.VirtualHost
	CacheAlgorithm cache.Algorithm
	Storage        storage.Storage
}

// New creates and returns a new VirtualHost struct with the specified parameters.
func New(config config.VirtualHost, cm cache.Algorithm, storage storage.Storage) *VirtualHost {
	return &VirtualHost{
		VirtualHost:    config,
		CacheAlgorithm: cm,
		Storage:        storage,
	}
}
