package types

import "net/http"

// Upstream represents an object that is used by the proxy handler for making
// requests to the configured upstream server or servers.
type Upstream interface {
	http.RoundTripper

	CancelRequest(*http.Request)

	GetAddress(string) (*UpstreamAddress, error)
}
