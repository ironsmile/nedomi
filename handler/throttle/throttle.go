package throttle

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"golang.org/x/net/context"

	"github.com/aybabtme/iocontrol"
	"github.com/ironsmile/nedomi/config"
	"github.com/ironsmile/nedomi/types"
)

// New creates and returns a ready to used ServerStatusHandler.
func New(cfg *config.Handler, l *types.Location, next types.RequestHandler) (types.RequestHandler, error) {
	var s struct {
		Speed types.BytesSize `json:"speed"`
	}
	if err := json.Unmarshal(cfg.Settings, &s); err != nil {
		return nil, fmt.Errorf("handler.throttle got error while parsing settings - %s", err)
	}
	if s.Speed == 0 {
		return nil, fmt.Errorf("handler.throttle needs to have speed settings > 0")
	}
	return types.RequestHandlerFunc(
		func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
			next.RequestHandle(ctx, &throttledResponseWriter{ResponseWriter: w,
				ThrottlerWriter: iocontrol.ThrottledWriter(w, int(s.Speed.Bytes()), time.Millisecond*10),
			}, r)
		}), nil
}

type throttledResponseWriter struct {
	http.ResponseWriter
	iocontrol.ThrottlerWriter
}

func (fw *throttledResponseWriter) Write(b []byte) (int, error) {
	return fw.ThrottlerWriter.Write(b)
}
