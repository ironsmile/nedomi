package types

import (
	"bytes"
	"time"
)

// App is an interface for an application to implement
type App interface {
	// Stats returns applicaiton wide stats
	Stats() AppStats

	// Started returns the time at which the app was started
	Started() time.Time

	// Version returns a string representation of the version of the app
	Version() AppVersion

	// GetLocationFor returns the Location that mathes the provided host and path
	GetLocationFor(host, path string) *Location
}

// AppStats are stats for the whole application
type AppStats struct {
	Requests, Responded, NotConfigured uint64
}

// AppVersion is struct representing an App version
type AppVersion struct {
	Dirty     bool
	Version   string
	GitHash   string
	GitTag    string
	BuildTime time.Time
}

func (a AppVersion) String() string {
	var ver = bytes.NewBufferString(a.Version)
	if a.GitTag != "" {
		ver.WriteRune('-')
		ver.WriteString(a.GitTag)
	} else if a.GitHash != "" {
		ver.WriteRune('-')
		ver.WriteString(a.GitHash)
	}
	if a.Dirty {
		ver.WriteString("-dirty")
	}

	ver.WriteString(" build at ")
	ver.WriteString(a.BuildTime.String())

	return ver.String()
}
