package upstream

import (
	// "github.com/gophergala/nedomi/types"
	"net/http"
)

type Upstream interface {
	GetRequestPartial(path string, start, end uint64) (*http.Response, error)

	GetSize(path string) (int64, error)

	GetHeader(path string) (http.Header, error)

	GetRequest(path string) (*http.Response, error)
}
