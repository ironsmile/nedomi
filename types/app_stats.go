package types

import (
	"strings"
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

	// GetUpstream gets an upstream by it's id, nil is returned if no such is defined
	GetUpstream(id string) Upstream
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
	ver := []string{a.Version}
	if a.GitTag != "" {
		ver = append(ver, "-", a.GitTag)
	} else if a.GitHash != "" {
		ver = append(ver, "-", a.GitHash)
	}
	if a.Dirty {
		ver = append(ver, "-dirty")
	}

	ver = append(ver, " build at ", a.BuildTime.String())

	return strings.Join(ver, "")
}
