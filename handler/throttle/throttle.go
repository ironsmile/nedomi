package throttle

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ironsmile/nedomi/config"
	"github.com/ironsmile/nedomi/contexts"
	"github.com/ironsmile/nedomi/types"
	"github.com/ironsmile/nedomi/utils"
	"github.com/ironsmile/nedomi/utils/httputils"
)

// Configuration is the struct the handler settings will be unmarshalled in
type Configuration struct {
	// Speed is the speed to which to throttle
	Speed types.BytesSize `json:"speed"`
}

// New creates and returns a ready to used ServerStatusHandler.
func New(cfg *config.Handler, l *types.Location, next http.Handler) (http.Handler, error) {
	if next == nil {
		return nil, types.NilNextHandler("throttle")
	}

	var c Configuration

	if err := json.Unmarshal(cfg.Settings, &c); err != nil {
		return nil, utils.ShowContextOfJSONError(err, cfg.Settings)
	}

	if c.Speed == 0 {
		return nil, fmt.Errorf("handler.throttle needs to have speed settings > 0")
	}

	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			conn, ok := contexts.GetConn(r.Context())
			if !ok {
				httputils.Error(w, http.StatusInternalServerError)
				return
			}
			conn.SetThrottle(c.Speed)
			next.ServeHTTP(w, r)
		}), nil
}
