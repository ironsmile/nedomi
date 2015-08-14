// Package upstream deals with connections to the virtual hosts' upstreams. It defines
// the Upstream interface. There is only one upstream object at the moment. It
// recongizes how to make the upstream request using the virtual host argument.
package upstream

import (
	"net/http"
)

// Upstream is an interface that is used by all implementations that deal with
// connections to the virtual hosts' upstreams.
type Upstream interface {
	GetRequestPartial(path string, start, end uint64) (*http.Response, error)

	GetSize(path string) (int64, error)

	GetHeader(path string) (*http.Response, error)

	GetRequest(path string) (*http.Response, error)
}
