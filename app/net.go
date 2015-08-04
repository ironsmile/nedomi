package app

import (
	"net/http"
	"strings"

	"github.com/ironsmile/nedomi/cache"
	"github.com/ironsmile/nedomi/handler"
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

	cancelCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	valueCtx := cache.NewContext(cancelCtx, app.cacheManagers)

	reqHandler.RequestHandle(valueCtx, writer, req, vh)
}
