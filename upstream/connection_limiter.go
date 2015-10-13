package upstream

import "net/http"

// NewConnectionLimiter creates a wrapper around the supplied RoundTripper that
// restricts the maximum number of concurrent requests through it.
func NewConnectionLimiter(base http.RoundTripper, limit uint32) http.RoundTripper {
	//!TODO: implement :)

	return base
}
