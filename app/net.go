package app

import (
	"net/http"
	"strings"

	"github.com/ironsmile/nedomi/contexts/vhost"
	"github.com/ironsmile/nedomi/types"
)

func (app *Application) findVirtualHost(r *http.Request) *types.VirtualHost {

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

	vh.Handler.RequestHandle(vhost.NewContext(app.ctx, vh), writer, req, vh)
}
