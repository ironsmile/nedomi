package mock

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/ironsmile/nedomi/config"
	"github.com/ironsmile/nedomi/types"

	"golang.org/x/net/context"
)

// DefaultResponseCode is the default response code for all requests.
const DefaultResponseCode = http.StatusOK

// DefaultResponse is the default response for all requests.
const DefaultResponse = "Hello"

// Handler is used to set custom responses for testing handler-related things.
type Handler struct {
	*http.ServeMux
}

// New creates and returns a new http.ServeMux instance that serves as a
// mock upstream. The default handler can be specified. If nil, a simple handler
// that always returns 200 and "Hello" is used.
func New(_ *config.Handler, _ *types.Location, _ types.RequestHandler) (*Handler, error) {
	return nil, errors.New("Should not create mock handlers via the normal interface")
}

// NewHandler creates and returns a new http.ServeMux instance that serves as a
// mock upstream. The default handler can be specified. If nil, a simple handler
// that always returns 200 and "Hello" is used.
func NewHandler(defaultHandler http.HandlerFunc) *Handler {
	handler := http.NewServeMux()
	if defaultHandler != nil {
		handler.HandleFunc("/", defaultHandler)
	} else {
		handler.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(DefaultResponseCode)
			fmt.Fprintf(w, DefaultResponse)
		})
	}

	return &Handler{handler}
}

// RequestHandle implements the interface
func (m *Handler) RequestHandle(_ context.Context, w http.ResponseWriter, r *http.Request) {
	m.ServeHTTP(w, r)
}
