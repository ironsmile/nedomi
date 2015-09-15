package upstream

import (
	"fmt"
	"net/http"
)

// MockDefaultResponseCode is the default response code for all requests.
const MockDefaultResponseCode = http.StatusOK

// MockDefaultResponse is the default response for all requests.
const MockDefaultResponse = "Hello"

// NewMock creates and returns a new http.ServeMux instance that serves as a
// mock upstream. The default handler can be specified. If nil, a simple handler
// that always returns 200 and "Hello" is used.
func NewMock(defaultHandler *http.HandlerFunc) *http.ServeMux {
	//!TODO: maybe use a custom type instaed of http.ServeMux directly. That way
	// we can add extra helper functions like delayed handling, responce copiers
	// and other goodies.

	upstream := http.NewServeMux()
	if defaultHandler != nil {
		upstream.HandleFunc("/", *defaultHandler)
	} else {
		upstream.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(MockDefaultResponseCode)
			fmt.Fprintf(w, MockDefaultResponse)
		})
	}

	return upstream
}
