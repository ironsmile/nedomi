package upstream

import (
	"net/http"
)

type Upstream interface {
	GetRequest(path string, start, end uint64) (*http.Response, error)
}
