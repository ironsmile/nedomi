package dir

import (
	"encoding/json"
	"fmt"
	"net/http"

	"golang.org/x/net/context"

	"github.com/ironsmile/nedomi/config"
	"github.com/ironsmile/nedomi/types"
)

// New creates and returns a ready to used ServerStatusHandler.
func New(cfg *config.Handler, l *types.Location, next types.RequestHandler) (types.RequestHandler, error) {
	var s struct {
		Root string `json:"root"`
	}
	if err := json.Unmarshal(cfg.Settings, &s); err != nil {
		return nil, fmt.Errorf("dir handler: error while parsing settings - %s", err)
	}

	fs := http.FileServer(http.Dir(s.Root))
	return types.RequestHandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		fs.ServeHTTP(w, r)
	}), nil
}
