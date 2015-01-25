package upstream

import (
	"github.com/gophergala/nedomi/config"
	"net/http"
)

type Upstream interface {
	GetRequestPartial(vh *config.VirtualHost, path string, start, end uint64) (*http.Response, error)

	GetSize(vh *config.VirtualHost, path string) (int64, error)

	GetHeader(vh *config.VirtualHost, path string) (http.Header, error)

	GetRequest(vh *config.VirtualHost, path string) (*http.Response, error)
}
