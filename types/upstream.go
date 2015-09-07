package types

import (
	"net/http"
)

// Upstream is an interface that is used by all implementations that deal with
// connections to the virtual hosts' upstreams.
type Upstream interface {
	ServeHTTP(http.ResponseWriter, *http.Request)

	//!TODO: method to get request statistics for the status page
}
