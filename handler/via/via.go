package via

import (
	"encoding/json"
	"net/http"

	"golang.org/x/net/context"

	"github.com/ironsmile/nedomi/config"
	"github.com/ironsmile/nedomi/types"
)

const defaultText = "nedomi 1.1"

// Via writes the via header
type Via struct {
	next types.RequestHandler
	text string
}

// RequestHandle writes the Via header to the http.ResponseWriter
func (v *Via) RequestHandle(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Via", v.text) //!TODO generalize it in static header adder ?
	v.next.RequestHandle(ctx, w, r)
}

// New creates and returns a ready to used ServerStatusHandler.
func New(cfg *config.Handler, l *types.Location, next types.RequestHandler) (*Via, error) {
	var text = defaultText
	if len(cfg.Settings) != 0 {
		var settings struct {
			Text string `json:"text"`
		}
		if err := json.Unmarshal(cfg.Settings, &settings); err != nil {
			return nil, err
		}
		text = settings.Text
	}

	return &Via{
		next: next,
		text: text,
	}, nil
}
