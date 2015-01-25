package upstream

import (
	"net/http"
)

type Upstream interface {
	GetRequest(path string, start, end uint64) (*http.Response, error)

	GetSize(path string) (uint64, error)
	GetHeader(path string) (http.Header, error)
}
