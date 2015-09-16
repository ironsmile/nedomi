package mp4

import (
	"fmt"
	"io"
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

	var rr = &rangeReader{ctx: ctx, req: copyRequest(r), location: l, next: m.next}
	var video = m.mapMp4File(rr)
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
	w.Header().Set("Content-Length", strconv.Itoa(int(cl.Size())))
	w.WriteHeader(http.StatusOK)
	size, err := cl.WriteTo(w)
	m.logger.Debugf("wrote %d", size)
	if err != nil {
		m.logger.Errorf("error on writing the clip response - %s", err)
	}
}

type rangeReader struct {
	ctx      context.Context
	req      *http.Request
	location *types.Location
	next     types.RequestHandler
}

func (rr *rangeReader) Range(start, length uint64) io.ReadCloser {
	newreq := copyRequest(rr.req)
	newreq.Header.Set("Range", utils.HTTPRange{Start: start, Length: length}.Range())
	var in, out = io.Pipe()
	flexible := utils.NewFlexibleResponseWriter(func(frw *utils.FlexibleResponseWriter) {
		if frw.Code != http.StatusPartialContent {
			out.CloseWithError(errUnsatisfactoryResponse)
		}
		frw.BodyWriter = out
	})
	go func() {
		defer out.Close()
		rr.next.RequestHandle(rr.ctx, flexible, newreq, rr.location)
	}()

	return in
}
func (rr *rangeReader) RangeRead(start, length uint64) (io.ReadCloser, error) {
	return rr.Range(start, length), nil
}

func (m *mp4Handler) mapMp4File(rr *rangeReader) *mp4.MP4 {
	video := &mp4.MP4{}
	var in = rr.Range(0, firstRequestSize)
	var currentOffset uint64
	var leftFromCurrentRequest uint64 = firstRequestSize

	for {
		var boxes []mp4.Box
		h, err := mp4.DecodeHeader(in)
		if err != nil {
			if err == io.EOF {
				return video
			}
			m.logger.Error(err)
			return nil
		}
		if h.Type == "mdat" { // don't decode that now
			m.logger.Debugf("got mdat(%d) at %d", h.Size, currentOffset)
			video.Mdat = &mp4.MdatBox{Offset: currentOffset, ContentSize: mp4.RemoveHeaderSize(h.Size)}
			if video.Moov != nil { // done
				return video
			}
			currentOffset += h.Size
			leftFromCurrentRequest = firstRequestSize
			in.Close()

			in = rr.Range(currentOffset, firstRequestSize)

			continue
		}
		requiredFromRequest := (h.Size + mp4.HeaderSizeFor(h.Size))
		if requiredFromRequest > leftFromCurrentRequest {
			var start, length = currentOffset + leftFromCurrentRequest, requiredFromRequest - leftFromCurrentRequest + firstRequestSize
			leftFromCurrentRequest += length
			in = utils.MultiReadCloser(in, rr.Range(start, length))
		}
		box, err := mp4.DecodeBox(h, in)
		if err != nil {
			m.logger.Error(err)
			return nil
		}
		switch h.Type {
		case "ftyp":
			m.logger.Debugf("got ftyp(%d) at %d", h.Size, currentOffset)
			video.Ftyp = box.(*mp4.FtypBox)
		case "moov":
			m.logger.Debugf("got moov(%d) at %d", h.Size, currentOffset)
			video.Moov = box.(*mp4.MoovBox)
		default:
			m.logger.Debugf("got a box(%d) at %d", h.Size, currentOffset)
			boxes = append(boxes, box) // not actually written in the output - hope it's not interesting :P
		}
		leftFromCurrentRequest -= h.Size
		currentOffset += h.Size
	}
}
