package pprof

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/pprof"
	"path"

	"golang.org/x/net/context"

	"github.com/ironsmile/nedomi/config"
	"github.com/ironsmile/nedomi/types"
)

// New creates and returns a ready to used ServerStatusHandler.
func New(cfg *config.Handler, l *types.Location, next types.RequestHandler) (types.RequestHandler, error) {
	var s = defaultSettings
	if err := json.Unmarshal(cfg.Settings, &s); err != nil {
		return nil, fmt.Errorf("error while parsing settings for handler.pprof - %s", err)
	}
	var mux = http.NewServeMux()
	mux.Handle(s.Path, http.HandlerFunc(pprof.Index))
	mux.Handle(path.Join(s.Path, "cmdline"), http.HandlerFunc(pprof.Cmdline))
	mux.Handle(path.Join(s.Path, "profile"), http.HandlerFunc(pprof.Profile))
	mux.Handle(path.Join(s.Path, "symbol"), http.HandlerFunc(pprof.Symbol))
	mux.Handle(path.Join(s.Path, "trace"), http.HandlerFunc(pprof.Trace))
	return types.RequestHandlerFunc(func(ctx context.Context, w http.ResponseWriter, req *http.Request, _ *types.Location) {
		mux.ServeHTTP(w, req)
	}), nil
}

var defaultSettings = settings{
	Path: "/debug/pprof/",
}

type settings struct {
	Path string `json:"path"`
}
