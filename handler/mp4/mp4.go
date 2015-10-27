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
	"github.com/ironsmile/nedomi/utils/httputils"
)

const (
	startKey = "start"
)

var errUnsatisfactoryResponse = fmt.Errorf("unsatisfactory response from the next handler")

// New creates and returns a ready to used ServerStatusHandler.
func New(cfg *config.Handler, loc *types.Location, next types.RequestHandler) (types.RequestHandler, error) {
	if next == nil {
		return nil, types.NilNextHandler("mp4")
	}

	// !TODO parse config
	return &mp4Handler{
		next: next,
		loc:  loc,
	}, nil
}

type mp4Handler struct {
	next types.RequestHandler
	loc  *types.Location
}

func copyRequest(r *http.Request) *http.Request {
	req := *r
	req.Header = http.Header{}
	url := *r.URL
	req.URL = &url
	httputils.CopyHeaders(r.Header, req.Header)
	return &req
}

func removeQueryArgument(u *url.URL, arguments ...string) {
	query := u.Query()
	for _, argument := range arguments {
		query.Del(argument)
	}
	u.RawQuery = query.Encode()
}

func (m *mp4Handler) RequestHandle(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	// Handle only GET requests with ContentLength of 0 without a Range header
	if r.Method != "GET" || len(r.Header.Get("Range")) > 0 || r.ContentLength > 0 {
		m.next.RequestHandle(ctx, w, r)
		return
	}

	// parse the request
	var start, err = strconv.Atoi(r.URL.Query().Get(startKey))
	if err != nil || 0 >= start { // that start is not ok
		m.next.RequestHandle(ctx, w, r)
		return
	}
	var startTime = time.Duration(start) * time.Second
	var newreq = copyRequest(r)
	removeQueryArgument(newreq.URL, startKey)
	var header = make(http.Header)
	var rr = &rangeReader{
		ctx:      ctx,
		req:      copyRequest(newreq),
		location: m.loc,
		next:     m.next,
		callback: func(frw *httputils.FlexibleResponseWriter) bool {
			if len(header) == 0 {
				httputils.CopyHeaders(frw.Header(), header, skipHeaders...)
			} else {
				return frw.Header().Get("Last-Modified") == header.Get("Last-Modified")
			}
			return true
		},
	}
	var video *mp4.MP4
	video, err = mp4.Decode(rr)
	if err != nil {
		m.loc.Logger.Errorf("error from the mp4.Decode - %s", err)
		m.next.RequestHandle(ctx, w, r)
		return
	}
	if video == nil || video.Moov == nil { // missing something?
		m.next.RequestHandle(ctx, w, r)
		return
	}

	cl, err := clip.New(video, startTime, rr)
	if err != nil {
		m.loc.Logger.Errorf("error while clipping a video(%+v) - %s", video, err)
		m.next.RequestHandle(ctx, w, r)
		return
	}
	httputils.CopyHeaders(header, w.Header())
	w.Header().Set("Content-Length", strconv.FormatUint(cl.Size(), 10))
	size, err := cl.WriteTo(w)
	m.loc.Logger.Debugf("wrote %d", size)
	if err != nil {
		m.loc.Logger.Errorf("error on writing the clip response - %s", err)
	}
	if uint64(size) != cl.Size() {
		m.loc.Logger.Errorf("handler.mp4[%p]: expected to write %d but wrote %d", m, cl.Size(), size)
	}
}

var skipHeaders = []string{
	"Content-Range",
}
