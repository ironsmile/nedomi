package httputils

import (
	"net/http"

	"github.com/ironsmile/nedomi/utils"
)

// CopyHeaders copies headers from `from` to `to` except for the `exceptions`
func CopyHeaders(from, to http.Header, exceptions ...string) {
	for k := range from {
		if !contains(exceptions, k) {
			to[k] = utils.CopyStringSlice(from[k])
		}
	}
}

func contains(heap []string, needle string) bool {
	for _, straw := range heap {
		if straw == needle {
			return true
		}
	}

	return false
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
