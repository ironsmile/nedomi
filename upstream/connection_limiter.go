package upstream

// newConnectionLimiter creates a wrapper around the supplied upClient that
// restricts the maximum number of concurrent requests through it.
func newConnectionLimiter(base upClient, limit uint32) upClient {
	//!TODO: implement :)

	return base
}
