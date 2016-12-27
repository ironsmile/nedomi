package types

import (
	"fmt"
	"net/http"

	"golang.org/x/net/context"
)

// RequestHandler interface defines the RequestHandle function. All nedomi handle
// modules must implement this interface.
type RequestHandler interface {

	// RequestHandle is function similar to the http.ServeHTTP. It differs only
	// in that it has context as an extra argument.
	RequestHandle(context.Context, http.ResponseWriter, *http.Request)
}

// The RequestHandlerFunc type is an adapter to allow the use of ordinary functions as request handlers.
// If f is a function with the appropriate signature, HandlerFunc(f) is a Handler object that calls f.
type RequestHandlerFunc func(context.Context, http.ResponseWriter, *http.Request)

// RequestHandle with rhf(ctx, w, req)
func (rhf RequestHandlerFunc) RequestHandle(ctx context.Context, w http.ResponseWriter, req *http.Request) {
	rhf(ctx, w, req)
}

// NilNextHandler is an error type to be returned when a next handler is required but it's nil
type NilNextHandler string

func (n NilNextHandler) Error() string {
	return fmt.Sprintf("%s: next handler is required but it's nil", n)
}

// NotNilNextHandler is an error type to be returned when a next handler is not nil but it should be
type NotNilNextHandler string

func (n NotNilNextHandler) Error() string {
	return fmt.Sprintf("%s: next handler is not nil but it should be", n)
}
