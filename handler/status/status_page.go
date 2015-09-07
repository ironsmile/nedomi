package status

import (
	"html/template"
	"net/http"

	"golang.org/x/net/context"

	"github.com/ironsmile/nedomi/contexts"
	"github.com/ironsmile/nedomi/types"
)

// ServerStatusHandler is a simple handler that handles the server status page.
type ServerStatusHandler struct {
}

// RequestHandle servers the status page.
//!TODO: Do not parse the template every request
func (ssh *ServerStatusHandler) RequestHandle(ctx context.Context,
	w http.ResponseWriter, r *http.Request, l *types.Location) {

	orchestrators, ok := contexts.GetStorageOrchestrators(ctx)
	if !ok {
		err := "Error: could not get the cache algorithm from the context!"
		l.Logger.Error(err)
		w.Write([]byte(err))
		return
	}

	tmpl, err := template.ParseFiles("handler/status/templates/status_page.html")

	if err != nil {
		l.Logger.Errorf("Error parsing template file: %s", err)
		w.Write([]byte(err.Error()))
		return
	}

	l.Logger.Logf("[%p] 200 Status page\n", r)
	w.WriteHeader(200)

	if err := tmpl.Execute(w, orchestrators); err != nil {
		w.Write([]byte(err.Error()))
	}

	return
}

// New creates and returns a ready to used ServerStatusHandler.
func New() *ServerStatusHandler {
	return &ServerStatusHandler{}
}
