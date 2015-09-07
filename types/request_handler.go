package types

import (
	"net/http"

	"golang.org/x/net/context"
)

// RequestHandler interface defines the RequestHandle funciton. All nedomi handle
// modules must implement this interface.
type RequestHandler interface {

	// RequestHandle is function similar to the http.ServeHTTP. It differs only in
	// that it has a context and a location as extra arguments.
	RequestHandle(context.Context, http.ResponseWriter, *http.Request, *Location)
}
