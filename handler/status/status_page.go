package status

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"path"

	"golang.org/x/net/context"

	"github.com/ironsmile/nedomi/config"
	"github.com/ironsmile/nedomi/contexts"
	"github.com/ironsmile/nedomi/types"
)

// ServerStatusHandler is a simple handler that handles the server status page.
type ServerStatusHandler struct {
	tmpl *template.Template
}

// RequestHandle servers the status page.
func (ssh *ServerStatusHandler) RequestHandle(ctx context.Context,
	w http.ResponseWriter, r *http.Request, l *types.Location) {

	cacheZones, ok := contexts.GetCacheZones(ctx)
	if !ok {
		err := "Error: could not get the cache algorithm from the context!"
		l.Logger.Error(err)
		w.Write([]byte(err))
		return
	}

	l.Logger.Logf("[%p] 200 Status page", r)
	w.WriteHeader(200)

	if err := ssh.tmpl.Execute(w, cacheZones); err != nil {
		w.Write([]byte(err.Error()))
	}

	return
}

// New creates and returns a ready to used ServerStatusHandler.
func New(cfg *config.Handler, l *types.Location, next types.RequestHandler) (*ServerStatusHandler, error) {
	var s = defaultSettings
	if err := json.Unmarshal(cfg.Settings, &s); err != nil {
		return nil, fmt.Errorf("error while parsing settings for handler.status - %s", err)
	}

	var statusFilePath = path.Join(s.Path, "status_page.html")
	var tmpl, err = template.ParseFiles(statusFilePath)
	if err != nil {
		return nil, fmt.Errorf("error on opening %s - %s", statusFilePath, err)
	}

	return &ServerStatusHandler{
		tmpl: tmpl,
	}, nil
}

var defaultSettings = serverStatusHandlerSettings{
	Path: "handler/status/templates",
}

type serverStatusHandlerSettings struct {
	Path string `json:"path"`
}
