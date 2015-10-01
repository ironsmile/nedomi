package types

import "net/url"

// Location links a config location to its cache algorithm and a storage object.
type Location struct {
	Name                  string
	CacheKey              string
	Handler               RequestHandler
	Cache                 *CacheZone //!TODO: this and the one below should be part of the cache handler settings
	Upstream              Upstream
	Logger                Logger
	CacheKeyIncludesQuery bool
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
