package status

import (
	"html/template"
	"log"
	"net/http"

	"golang.org/x/net/context"

	"github.com/ironsmile/nedomi/cache"
	"github.com/ironsmile/nedomi/vhost"
)

// ServerStatusHandler is a simple handler that handles the server status page.
type ServerStatusHandler struct {
}

// RequestHandle servers the status page.
//!TODO: Do not parse the template every request
func (ssh *ServerStatusHandler) RequestHandle(ctx context.Context,
	w http.ResponseWriter, r *http.Request, vh *vhost.VirtualHost) {

	cms, ok := cache.FromContext(ctx)
	if !ok {
		err := "Error: could not get the cache algorithm from the context!"
		log.Printf(err)
		w.Write([]byte(err))
		return
	}

	tmpl, err := template.ParseFiles("handler/status/templates/status_page.html")

	if err != nil {
		log.Printf("Error parsing template file: %s", err)
		w.Write([]byte(err.Error()))
		return
	}

	log.Printf("[%p] 200 Status page\n", r)
	w.WriteHeader(200)

	if err := tmpl.Execute(w, cms); err != nil {
		w.Write([]byte(err.Error()))
	}

	return
}

// New creates and returns a ready to used ServerStatusHandler.
func New() *ServerStatusHandler {
	return &ServerStatusHandler{}
}
