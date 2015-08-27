package types

import "net/url"

// VirtualHost links a config vritual host to its cache algorithm and a storage object.
type VirtualHost struct {
	Name            string
	CacheKey        string
	Handler         RequestHandler
	Orchestrator    StorageOrchestrator
	UpstreamAddress *url.URL //!TODO: remove, this should not be needed
	Upstream        Upstream
	Logger          Logger
}
