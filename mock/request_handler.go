package mock

import (
	"fmt"
	"net/http"

	"golang.org/x/net/context"
)

// DefaultRequestHandlerResponseCode is the default response code for all requests.
const DefaultRequestHandlerResponseCode = http.StatusOK

// DefaultRequestHandlerResponse is the default response for all requests.
const DefaultRequestHandlerResponse = "Hello"

// RequestHandler is used to set custom responses for testing handler-related things.
type RequestHandler struct {
	*http.ServeMux
}

// NewRequestHandler creates and returns a new http.ServeMux instance that serves as a
// mock upstream. The default handler can be specified. If nil, a simple handler
// that always returns 200 and "Hello" is used.
func NewRequestHandler(defaultHandler http.HandlerFunc) *RequestHandler {
	handler := http.NewServeMux()
	if defaultHandler != nil {
		handler.HandleFunc("/", defaultHandler)
	} else {
		handler.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(DefaultRequestHandlerResponseCode)
			fmt.Fprintf(w, DefaultRequestHandlerResponse)
		})
	}

	return &RequestHandler{handler}
}

// RequestHandle implements the interface
func (m *RequestHandler) RequestHandle(_ context.Context, w http.ResponseWriter, r *http.Request) {
	m.ServeHTTP(w, r)
}
