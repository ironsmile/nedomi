package status

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path"
	"strings"

	"golang.org/x/net/context"

	"github.com/ironsmile/nedomi/config"
	"github.com/ironsmile/nedomi/contexts"
	"github.com/ironsmile/nedomi/types"
	"github.com/ironsmile/nedomi/utils/testutils"
)

// ServerStatusHandler is a simple handler that handles the server status page.
type ServerStatusHandler struct {
	tmpl *template.Template
}

// RequestHandle servers the status page.
func (ssh *ServerStatusHandler) RequestHandle(ctx context.Context,
	w http.ResponseWriter, r *http.Request, l *types.Location) {

	app, ok := contexts.GetApp(ctx)
	if !ok {
		err := "Error: could not get the App from the context!"
		if _, writeErr := w.Write([]byte(err)); writeErr != nil {
			l.Logger.Errorf("error while writing error to client: `%s`; Original error `%s`", writeErr, err)
		} else {
			l.Logger.Error(err)
		}
		return
	}

	cacheZones, ok := contexts.GetCacheZones(ctx)
	if !ok {
		err := "Error: could not get the cache zones from the context!"
		if _, writeErr := w.Write([]byte(err)); writeErr != nil {
			l.Logger.Errorf("error while writing error to client: `%s`; Original error `%s`", writeErr, err)
		} else {
			l.Logger.Error(err)
		}
		return
	}

	l.Logger.Logf("[%p] 200 Status page", r)
	w.WriteHeader(200)

	if err := ssh.tmpl.Execute(w, newStatistics(app.Stats(), cacheZones)); err != nil {
		if _, writeErr := w.Write([]byte(err.Error())); writeErr != nil {
			l.Logger.Errorf("error while writing error to client: `%s`; Original error `%s`", writeErr, err)
		}
	}

	return
}

func newStatistics(appStats types.AppStats, cacheZones map[string]*types.CacheZone) statisticsRoot {
	return statisticsRoot{
		Requests:      appStats.Requests,
		Responded:     appStats.Responded,
		NotConfigured: appStats.NotConfigured,
		InFlight:      appStats.Requests - appStats.Responded - appStats.NotConfigured,
		CacheZones:    cacheZones,
	}
}

type statisticsRoot struct {
	Requests, Responded, NotConfigured, InFlight uint64
	CacheZones                                   map[string]*types.CacheZone
}

// New creates and returns a ready to used ServerStatusHandler.
func New(cfg *config.Handler, l *types.Location, next types.RequestHandler) (*ServerStatusHandler, error) {
	var s = defaultSettings
	if err := json.Unmarshal(cfg.Settings, &s); err != nil {
		return nil, fmt.Errorf("error while parsing settings for handler.status - %s", err)
	}

	// In case of:
	//  * the path is missing and it is relative
	//		or
	//  * the path is not a directory
	// we try to guess the project's root and use s.Path as a relative to it
	// one in hope it will match the templates' directory.
	if st, err := os.Stat(s.Path); (err != nil && err.(*os.PathError) != nil &&
		!strings.HasPrefix(s.Path, "/")) || (err == nil && !st.IsDir()) {

		projPath, err := testutils.ProjectPath()
		if err == nil {
			fullPath := path.Join(projPath, s.Path)
			if st, err := os.Stat(fullPath); err == nil && st.IsDir() {
				s.Path = fullPath
			}
		}
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
