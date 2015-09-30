package types

// App is an interface for an application to implement
type App interface {
	// returns applicaiton wide stats
	Stats() AppStats
}

// AppStats are stats for the whole application
type AppStats struct {
	Requests, Responded, NotConfigured uint64
}
