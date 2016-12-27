package app

import (
	"encoding/hex"
	"net/http"
	"strings"

	"github.com/ironsmile/nedomi/contexts"
	"github.com/ironsmile/nedomi/types"
	"github.com/ironsmile/nedomi/utils/httputils"
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
	var (
		reqID    = app.newRequestIDFor(app.stats.requested())
		ctx      = contexts.NewIDContext(app.ctx, reqID)
		location = app.GetLocationFor(req.Host, req.URL.Path)
	)

	if location == nil || location.Handler == nil {
		req = req.WithContext(ctx)
		defer app.stats.notConfigured()
		app.notConfiguredHandler.ServeHTTP(writer, req)
		return
	}

	defer app.stats.responded()

	var conn, ok = app.conns.find(req.RemoteAddr)
	if !ok { // highly unlikely
		app.GetLogger().Errorf("couldn't find connection for req with addr %s!%s!%s\n",
			req.RemoteAddr, reqID, req.URL.Path)
		httputils.Error(writer, http.StatusInternalServerError)
		return
	}

	ctx = contexts.NewConnContext(ctx, conn) // TODO: figure out how to remove this
	req = req.WithContext(ctx)
	location.Handler.ServeHTTP(writer, req)
}

func newNotConfiguredHandler() http.Handler {
	return http.HandlerFunc(http.NotFound)
}

func (app *Application) newRequestIDFor(c uint64) types.RequestID {
	app.RLock()
	var appIdlen = len(app.cfg.ApplicationID)
	app.RUnlock()
	var reqID = types.RequestID(make([]byte, appIdlen+16))
	var last8 = reqID[appIdlen+8:]
	var t = app.started.Unix()
	copy(reqID, app.cfg.ApplicationID)
	// time is littleEndian, count is bigEndian and than or them
	last8[0] = byte(c>>56) | byte(t)
	last8[1] = byte(c>>48) | byte(t>>8)
	last8[2] = byte(c>>40) | byte(t>>16)
	last8[3] = byte(c>>32) | byte(t>>24)
	last8[4] = byte(c>>24) | byte(t>>32)
	last8[5] = byte(c>>16) | byte(t>>40)
	last8[6] = byte(c>>8) | byte(t>>48)
	last8[7] = byte(c) | byte(t>>56)

	hex.Encode(reqID[appIdlen:], last8)
	return reqID
}
