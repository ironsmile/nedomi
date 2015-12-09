package throttle

import (
	"encoding/json"
	"fmt"
	"net/http"

	"golang.org/x/net/context"

	"github.com/ironsmile/nedomi/config"
	"github.com/ironsmile/nedomi/contexts"
	"github.com/ironsmile/nedomi/types"
	"github.com/ironsmile/nedomi/utils"
)

// Configuration is the struct the handler settings will be unmarshalled in
type Configuration struct {
	// Speed is the speed to which to throttle
	Speed types.BytesSize `json:"speed"`
}

// New creates and returns a ready to used ServerStatusHandler.
func New(cfg *config.Handler, l *types.Location, next types.RequestHandler) (types.RequestHandler, error) {
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

	return types.RequestHandlerFunc(
		func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
			conn, _ := contexts.GetConn(ctx)
			conn.SetThrottle(c.Speed)
			next.RequestHandle(ctx, w, r)
		}), nil
}
