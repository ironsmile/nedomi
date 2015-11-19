package app

import (
	"encoding/hex"
	"net/http"
	"strings"

	"golang.org/x/net/context"

	"github.com/ironsmile/nedomi/contexts"
	"github.com/ironsmile/nedomi/types"
)

// GetLocationFor returns the Location that mathes the provided host and path
func (app *Application) GetLocationFor(host, path string) *types.Location {
	app.RLock()
	split := strings.Split(host, ":")
	vh, ok := app.virtualHosts[split[0]]
	app.RUnlock()
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
	var id = app.newIDFor(app.stats.requested())
	// new request
	location := app.GetLocationFor(req.Host, req.URL.Path)
	var ctx = contexts.NewIDContext(app.ctx, id)

	if location == nil || location.Handler == nil {
		defer app.stats.notConfigured()
		app.notConfiguredHandler.RequestHandle(ctx, writer, req)
		return
	}
	// location matched
	// stuff before the request is handled
	defer app.stats.responded()
	location.Handler.RequestHandle(ctx, writer, req)
	// after request is handled
}

func newNotConfiguredHandler() types.RequestHandler {
	return types.RequestHandlerFunc(func(_ context.Context, w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	})
}

func (app *Application) newIDFor(c uint64) types.ID {
	var appIdlen = len(app.cfg.ApplicationID)
	var result = types.ID(make([]byte, appIdlen+16))
	var last8 = result[appIdlen+8:]
	var t = app.started.Unix()
	copy(result, app.cfg.ApplicationID)
	// time is littleEndian, count is bigEndian and than or them
	last8[0] = byte(c>>56) | byte(t)
	last8[1] = byte(c>>48) | byte(t>>8)
	last8[2] = byte(c>>40) | byte(t>>16)
	last8[3] = byte(c>>32) | byte(t>>24)
	last8[4] = byte(c>>24) | byte(t>>32)
	last8[5] = byte(c>>16) | byte(t>>40)
	last8[6] = byte(c>>8) | byte(t>>48)
	last8[7] = byte(c) | byte(t>>56)

	hex.Encode(result[appIdlen:], last8)
	return result
}
