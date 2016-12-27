package dir

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ironsmile/nedomi/config"
	"github.com/ironsmile/nedomi/types"
)

// New creates and returns a ready to used ServerStatusHandler.
func New(cfg *config.Handler, l *types.Location, next http.Handler) (http.Handler, error) {
	var s struct {
		Root  string `json:"root"`
		Strip string `json:"strip"`
	}
	if err := json.Unmarshal(cfg.Settings, &s); err != nil {
		return nil, fmt.Errorf("dir handler: error while parsing settings - %s", err)
	}

	h := http.StripPrefix(s.Strip, http.FileServer(http.Dir(s.Root)))
	return h, nil
}
