package types

// App is an interface for an application to implement
type App interface {
	// Stats returns applicaiton wide stats
	Stats() AppStats

	// GetLocationFor returns the Location that mathes the provided host and path
	GetLocationFor(host, path string) *Location
}

// AppStats are stats for the whole application
type AppStats struct {
	Requests, Responded, NotConfigured uint64
}
