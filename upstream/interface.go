package upstream

import (
	// "github.com/gophergala/nedomi/types"
	"net/http"
)

type Upstream interface {
	GetRequestPartial(path string, start, end uint64) (*http.Response, error)

	GetSize(path string) (uint64, error)
	GetHeader(path string) (http.Header, error)
}
