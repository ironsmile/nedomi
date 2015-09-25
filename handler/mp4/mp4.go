package mp4

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"golang.org/x/net/context"

	"github.com/MStoykov/mp4"
	"github.com/MStoykov/mp4/clip"

	"github.com/ironsmile/nedomi/config"
	"github.com/ironsmile/nedomi/types"
	"github.com/ironsmile/nedomi/utils"
)

const (
	startKey         = "start"
	firstRequestSize = 4096
)

var errUnsatisfactoryResponse = fmt.Errorf("unsatisfactory response from the next handler")

// New creates and returns a ready to used ServerStatusHandler.
func New(cfg *config.Handler, l *types.Location, next types.RequestHandler) (types.RequestHandler, error) {
	// !TODO parse config
	return &mp4Handler{
		next:   next,
		logger: l.Logger,
	}, nil
}

type mp4Handler struct {
	next   types.RequestHandler
	logger types.Logger
}

func copyRequest(r *http.Request) *http.Request {
	req := *r
	req.Header = http.Header{}
	url := *r.URL
	req.URL = &url
	utils.CopyHeadersWithout(r.Header, req.Header)
	return &req
}

func removeQueryArgument(u *url.URL, arguments ...string) {
	query := u.Query()
	for _, argument := range arguments {
		query.Del(argument)
	}
	u.RawQuery = query.Encode()
}

func (m *mp4Handler) RequestHandle(ctx context.Context, w http.ResponseWriter, r *http.Request, l *types.Location) {
	// Handle only GET requests with ContentLength of 0 without a Range header
	if r.Method != "GET" || len(r.Header.Get("Range")) > 0 || r.ContentLength > 0 {
		m.next.RequestHandle(ctx, w, r, l)
		return
	}

	// parse the request
	var start, err = strconv.Atoi(r.URL.Query().Get(startKey))
	if err != nil || 0 >= start { // that start is not ok
		m.next.RequestHandle(ctx, w, r, l)
		return
	}
	var startTime = time.Duration(start) * time.Second
	var newreq = copyRequest(r)
	removeQueryArgument(newreq.URL, startKey)

	var rr = &rangeReader{ctx: ctx, req: copyRequest(newreq), location: l, next: m.next}
	var video *mp4.MP4
	video, err = mp4.Decode(rr)
	if err != nil {
		m.logger.Errorf("error from the mp4.Decode - %s", err)
		m.next.RequestHandle(ctx, w, r, l)
		return
	}
	if video == nil || video.Moov == nil { // missing something?
		m.next.RequestHandle(ctx, w, r, l)
		return
	}

	cl, err := clip.New(video, startTime, rr)
	if err != nil {
		m.logger.Errorf("error while clipping a video(%+v) - %s", video, err)
		m.next.RequestHandle(ctx, w, r, l)
		return
	}
	w.Header().Set("Content-Type", "video/mp4") // copy it from next
	w.Header().Set("Content-Length", strconv.FormatUint(cl.Size(), 10))
	w.WriteHeader(http.StatusOK)
	size, err := cl.WriteTo(w)
	m.logger.Debugf("wrote %d", size)
	if err != nil {
		m.logger.Errorf("error on writing the clip response - %s", err)
	}
}
