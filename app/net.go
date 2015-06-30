package app

import (
	"net/http"
	"strings"

	"github.com/ironsmile/nedomi/handler"
	"github.com/ironsmile/nedomi/vhost"
)

func (app *Application) findVirtualHost(r *http.Request) (*vhost.VirtualHost,
	handler.RequestHandler) {

	split := strings.Split(r.Host, ":")
	vhPair, ok := app.virtualHosts[split[0]]

	if !ok {
		return nil, nil
	}

	return vhPair.vhostStruct, vhPair.vhostHandler
}

func (app *Application) ServeHTTP(writer http.ResponseWriter, req *http.Request) {

	vh, reqHandler := app.findVirtualHost(req)

	if vh == nil || reqHandler == nil {
		http.NotFound(writer, req)
		return
	}

	reqHandler.RequestHandle(writer, req, vh)
}
