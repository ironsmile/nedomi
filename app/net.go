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
	vh := app.findVirtualHost(req)
	if vh == nil {
		http.NotFound(writer, req)
		return
	}

	location := vh.Muxer.Match(req.URL.Path)
	if location == nil {
		if vh.Handler != nil { //!TODO should this be this way ?
			vh.Handler.RequestHandle(contexts.NewLocationContext(app.ctx, &vh.Location), writer, req, &vh.Location)
		} else {
			http.NotFound(writer, req)
		}
		return
	}

	//!TODO change it to send the Location settings
	location.Handler.RequestHandle(contexts.NewLocationContext(app.ctx, location), writer, req, location)
}
