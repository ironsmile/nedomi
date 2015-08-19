package app

import (
	"net/http"
	"strings"

	"github.com/ironsmile/nedomi/handler"
	"github.com/ironsmile/nedomi/storage"
	"github.com/ironsmile/nedomi/vhost"

	"golang.org/x/net/context"
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

	//!TODO: create the background context once in the app init
	cancelCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//!TODO: move this to app.init as well, here we should add vhost in context
	valueCtx := storage.NewContext(cancelCtx, app.storages)

	reqHandler.RequestHandle(valueCtx, writer, req, vh)
}
