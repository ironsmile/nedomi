package types

// Location links a config location to its cache algorithm and a storage object.
type Location struct {
	Name         string
	CacheKey     string
	Handler      RequestHandler
	Orchestrator StorageOrchestrator
	Upstream     Upstream
	Logger       Logger
}

func (l *Location) String() string {
	return l.Name
}
