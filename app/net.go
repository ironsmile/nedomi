package app

import (
	"net/http"
	"strings"

	"github.com/ironsmile/nedomi/contexts"
)

func (app *Application) findVirtualHost(r *http.Request) *VirtualHost {
	split := strings.Split(r.Host, ":")
	vh, ok := app.virtualHosts[split[0]]

	if !ok {
		return nil
	}

	return vh
}

func (app *Application) ServeHTTP(writer http.ResponseWriter, req *http.Request) {
	app.stats.requested()
	// new request
	vh := app.findVirtualHost(req)
	if vh == nil {
		app.stats.notConfigured()
		http.NotFound(writer, req)
		return
	}

	// virtualHost found
	location := vh.Muxer.Match(req.URL.Path)
	if location == nil {
		if vh.Handler == nil {
			app.stats.notConfigured()
			http.NotFound(writer, req)
		} else {
			// do stuff before request is handled
			vh.Handler.RequestHandle(contexts.NewLocationContext(app.ctx, &vh.Location), writer, req, &vh.Location)
			// after request is handled
			app.stats.responded()
		}
		return
	}

	// location matched
	// stuff before the request is handled
	location.Handler.RequestHandle(contexts.NewLocationContext(app.ctx, location), writer, req, location)
	app.stats.responded()

	// after request is handled
}
