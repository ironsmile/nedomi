package headers

import (
	"net/http"

	"github.com/ironsmile/nedomi/config"
)

// headersRewrite rewrites headers
type headersRewrite config.HeadersRewrite

func (hr *headersRewrite) isEmpty() bool {
	return len(hr.RemoveHeaders) == 0 &&
		len(hr.AddHeaders) == 0 &&
		len(hr.SetHeaders) == 0
}

func (hr *headersRewrite) rewrite(headers http.Header) {
	for _, key := range hr.RemoveHeaders {
		headers.Del(key)
	}

	for key, values := range hr.AddHeaders {
		addValues(headers, key, values)
	}
	for key, values := range hr.SetHeaders {
		headers.Del(key)
		addValues(headers, key, values)
	}
}

func addValues(headers http.Header, key string, values []string) {
	for _, value := range values {
		headers.Add(key, value)
	}
}
