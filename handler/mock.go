package handler

import (
	"fmt"
	"net/http"

	"github.com/ironsmile/nedomi/types"
	"golang.org/x/net/context"
)

// MockDefaultResponseCode is the default response code for all requests.
const MockDefaultResponseCode = http.StatusOK

// MockDefaultResponse is the default response for all requests.
const MockDefaultResponse = "Hello"

// MockHandler is used to set custom responses for testing handler-related things.
type MockHandler struct {
	*http.ServeMux
}

// NewMock creates and returns a new http.ServeMux instance that serves as a
// mock upstream. The default handler can be specified. If nil, a simple handler
// that always returns 200 and "Hello" is used.
func NewMock(defaultHandler http.HandlerFunc) *MockHandler {
	//!TODO: maybe use a custom type instaed of http.ServeMux directly. That way
	// we can add extra helper functions like delayed handling, responce copiers
	// and other goodies.

	handler := http.NewServeMux()
	if defaultHandler != nil {
		handler.HandleFunc("/", defaultHandler)
	} else {
		handler.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(MockDefaultResponseCode)
			fmt.Fprintf(w, MockDefaultResponse)
		})
	}

	return &MockHandler{handler}
}

// RequestHandle implements the interface
func (m *MockHandler) RequestHandle(_ context.Context, w http.ResponseWriter, r *http.Request, _ *types.Location) {
	m.ServeHTTP(w, r)
}
