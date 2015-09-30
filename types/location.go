package types

// Location links a config location to its cache algorithm and a storage object.
type Location struct {
	Name     string
	CacheKey string
	Handler  RequestHandler
	Cache    *CacheZone //!TODO: this and the one below should be part of the cache handler settings
	Upstream Upstream
	Logger   Logger
}

func (l *Location) String() string {
	return l.Name
}
