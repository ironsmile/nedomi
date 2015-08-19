package types

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
