package storage

import (
	"github.com/ironsmile/nedomi/config"
	"github.com/ironsmile/nedomi/types"
)

// Orchestrator is responsible for coordinating and synchronizing the operations
// between the Storage, CacheAlgorithm and each vhost's Upstream server.
type Orchestrator struct {
	cfg       *config.CacheZone
	storage   types.Storage
	algorithm types.CacheAlgorithm
	logger    types.Logger
}
