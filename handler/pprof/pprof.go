package pprof

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/pprof"
	"path"

	"github.com/ironsmile/nedomi/config"
	"github.com/ironsmile/nedomi/types"
	"github.com/ironsmile/nedomi/utils"
)

var prefixToHandler = map[string]http.HandlerFunc{
	"cmdline": pprof.Cmdline,
	"profile": pprof.Profile,
	"symbol":  pprof.Symbol,
}

// New creates and returns a ready to used ServerStatusHandler.
func New(cfg *config.Handler, l *types.Location, next http.Handler) (http.Handler, error) {
	var s = defaultSettings
	if len(cfg.Settings) > 0 {
		if err := json.Unmarshal(cfg.Settings, &s); err != nil {
			return nil, fmt.Errorf("error while parsing settings for handler.pprof - %s",
				utils.ShowContextOfJSONError(err, cfg.Settings))
		}
	}
	var mux = http.NewServeMux()
	mux.HandleFunc(s.Path, pprof.Index)
	for prefix, handler := range prefixToHandler {
		mux.HandleFunc(path.Join(s.Path, prefix), handler)
	}
	return mux, nil
}

var defaultSettings = settings{
	Path: "/debug/pprof/",
}

type settings struct {
	Path string `json:"path"`
}
