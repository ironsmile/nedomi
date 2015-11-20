package types

import (
	"net/http"

	"golang.org/x/net/context"
)

// Upstream represents an object that is used by the proxy handler for making
// requests to the configured upstream server or servers.
type Upstream interface {
	Do(ctx context.Context, req *http.Request) (*http.Response, error)

	CancelRequest(*http.Request)

	GetAddress(string) (*UpstreamAddress, error)
}
