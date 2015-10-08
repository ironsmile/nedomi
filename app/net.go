package app

import (
	"net/http"
	"strings"

	"golang.org/x/net/context"

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
		app.notConfiguredHandler.RequestHandle(app.ctx, writer, req)
		return
	}
	// location matched
	// stuff before the request is handled
	defer app.stats.responded()
	location.Handler.RequestHandle(app.ctx, writer, req)
	// after request is handled
}

func newNotConfiguredHandler() types.RequestHandler {
	return types.RequestHandlerFunc(func(_ context.Context, w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	})
}
