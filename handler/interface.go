// Package handler deals with the main HTTP handler modules for nedomi. It describes the
// RequestHandler interface. Every subpackage *must* have a type which implements it.
package handler

import (
	"net/http"

	"github.com/ironsmile/nedomi/vhost"
)

// RequestHandler interface defines the RequestHandle funciton. All nedomi handle
// modules must implement this interface.
type RequestHandler interface {

	// RequestHandle is function similar to the http.ServeHTTP. It differs only in
	// that it has a vhost as a third argument.
	RequestHandle(http.ResponseWriter, *http.Request, *vhost.VirtualHost)
}
