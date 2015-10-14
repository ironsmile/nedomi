package httputils

import (
	"net/http"

	"github.com/ironsmile/nedomi/utils"
)

// CopyHeaders copies all headers from `from` to `to`.
func CopyHeaders(from, to http.Header) {
	for k := range from {
		to[k] = utils.CopyStringSlice(from[k])
	}
}

// CopyHeadersWithout copies headers from `from` to `to` except for the `exceptions`
func CopyHeadersWithout(from, to http.Header, exceptions ...string) {
	for k := range from {
		shouldCopy := true
		for _, e := range exceptions {
			if e == k {
				shouldCopy = false
				break
			}
		}
		if shouldCopy {
			to[k] = utils.CopyStringSlice(from[k])
		}
	}
}

// GetHopByHopHeaders returns a list of hop-by-hop headers. These should be
// removed when sending proxied responses to the client.
// http://www.w3.org/Protocols/rfc2616/rfc2616-sec13.html
func GetHopByHopHeaders() []string {
	return []string{
		"Connection",
		"Keep-Alive",
		"Proxy-Authenticate",
		"Proxy-Authorization",
		"Te", // canonicalized version of "TE"
		"Trailers",
		"Transfer-Encoding",
		"Upgrade",
	}
}
