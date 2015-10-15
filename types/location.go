package types

import (
	"net/url"
	"time"
)

// Location links a config location to its cache algorithm and a storage object.
type Location struct {
	Name                  string
	Handler               RequestHandler
	CacheKey              string
	CacheDefaultDuration  time.Duration
	CacheKeyIncludesQuery bool
	Cache                 *CacheZone //!TODO: move to the cache handler settings (plus all Cache* settings)
	Upstream              Upstream
	Logger                Logger
}

func (l *Location) String() string {
	return l.Name
}

// NewObjectIDForURL returns new ObjectID from the provided URL
func (l *Location) NewObjectIDForURL(u *url.URL) *ObjectID {
	if l.CacheKeyIncludesQuery {
		return NewObjectID(l.CacheKey, u.String())
	}
	return NewObjectID(l.CacheKey, u.Path)
}
