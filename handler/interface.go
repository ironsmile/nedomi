// Package handler deals with the main HTTP handler modules for nedomi. It describes the
// RequestHandler interface. Every subpackage *must* have a type which implements it.
package handler

import (
	"net/http"

	"golang.org/x/net/context"

	"github.com/ironsmile/nedomi/vhost"
)

// RequestHandler interface defines the RequestHandle funciton. All nedomi handle
// modules must implement this interface.
type RequestHandler interface {

	// RequestHandle is function similar to the http.ServeHTTP. It differs only in
	// that it has a context and a vhost as extra arguments.
	RequestHandle(context.Context, http.ResponseWriter, *http.Request, *vhost.VirtualHost)
}
