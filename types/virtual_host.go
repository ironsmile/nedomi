package types

// VirtualHost links a config vritual host to its cache algorithm and a storage object.
type VirtualHost struct {
	Name         string
	CacheKey     string
	Handler      RequestHandler
	Orchestrator StorageOrchestrator
	Upstream     Upstream
	Logger       Logger
}
