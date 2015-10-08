package types

import (
	"net/http"

	"golang.org/x/net/context"
)

// RequestHandler interface defines the RequestHandle funciton. All nedomi handle
// modules must implement this interface.
type RequestHandler interface {

	// RequestHandle is function similar to the http.ServeHTTP. It differs only
	// in that it has context as an extra argument.
	RequestHandle(context.Context, http.ResponseWriter, *http.Request)
}

// The RequestHandlerFunc type is an adapter to allow the use of ordinary functions as request handlers.
// If f is a function with the appropriate signature, HandlerFunc(f) is a Handler object that calls f.
type RequestHandlerFunc func(context.Context, http.ResponseWriter, *http.Request)

// RequestHandle with rhf(ctx, w, req, l)
func (rhf RequestHandlerFunc) RequestHandle(ctx context.Context, w http.ResponseWriter, req *http.Request) {
	rhf(ctx, w, req)
}
