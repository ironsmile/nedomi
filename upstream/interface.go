/*
   Package upstream deals with connections to the virtual hosts' upstreams. It defines
   the Upstream interface. There is only one upstream object at the moment. It
   recongizes how to make the upstream request using the virtual host argument.
*/
package upstream

import (
	"net/http"

	"github.com/ironsmile/nedomi/config"
)

type Upstream interface {
	GetRequestPartial(vh *config.VirtualHost, path string, start, end uint64) (*http.Response, error)

	GetSize(vh *config.VirtualHost, path string) (int64, error)

	GetHeader(vh *config.VirtualHost, path string) (http.Header, error)

	GetRequest(vh *config.VirtualHost, path string) (*http.Response, error)
}
