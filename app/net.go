package app

import (
	"net/http"
	"strings"

	"github.com/ironsmile/nedomi/contexts"
	"github.com/ironsmile/nedomi/types"
)

// GetLocationFor returns the Location that mathes the provided host and path
func (app *Application) GetLocationFor(host, path string) *types.Location {
	split := strings.Split(host, ":")
	vh, ok := app.virtualHosts[split[0]]
	if !ok {
		return nil
	}

	location := vh.Muxer.Match(path)
	if location == nil {
		return &vh.Location
	}
	return location
}

func (app *Application) ServeHTTP(writer http.ResponseWriter, req *http.Request) {
	app.stats.requested()
	// new request
	location := app.GetLocationFor(req.Host, req.URL.Path)

	if location == nil || location.Handler == nil {
		defer app.stats.notConfigured()
		http.NotFound(writer, req)
		return
	}
	// location matched
	// stuff before the request is handled
	defer app.stats.responded()
	location.Handler.RequestHandle(contexts.NewLocationContext(app.ctx, location), writer, req, location)
	// after request is handled
}
