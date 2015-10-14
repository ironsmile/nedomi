package upstream

// newConnectionLimiter creates a wrapper around the supplied RoundTripper that
// restricts the maximum number of concurrent requests through it.
func newConnectionLimiter(base upTransport, limit uint32) upTransport {
	//!TODO: implement :)

	return base
}
