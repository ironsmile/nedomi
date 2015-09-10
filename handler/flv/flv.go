package flv

import (
	"fmt"
	"net/http"
	"strconv"

	"golang.org/x/net/context"

	"github.com/ironsmile/nedomi/config"
	"github.com/ironsmile/nedomi/types"
)

var flvHeader = [13]byte{'F', 'L', 'V', 1, 5, 0, 0, 0, 9, 0, 0, 0, 0}

const startKey = "start"

// New creates and returns a ready to used ServerStatusHandler.
func New(cfg *config.Handler, l *types.Location, next types.RequestHandler) (types.RequestHandler, error) {
	return types.RequestHandlerFunc(
		func(ctx context.Context, w http.ResponseWriter, r *http.Request, l *types.Location) {
			var start, err = strconv.Atoi(r.URL.Query().Get(startKey))
			if err != nil || 0 >= start { // pass
				next.RequestHandle(ctx, w, r, l)
				return
			}
			r.URL.Query().Del(startKey) // clean that
			r.Header.Add("Range", fmt.Sprintf("bytes=%d-", start))
			next.RequestHandle(ctx, &flvWriter{w: w}, r, l)

		}), nil
}

type flvWriter struct {
	w             http.ResponseWriter
	headerWritten bool
	status        int
}

func (fw *flvWriter) Header() http.Header {
	return fw.w.Header()
}

func (fw *flvWriter) Write(b []byte) (int, error) {
	if !fw.headerWritten {
		fw.headerWritten = true
		if fw.status == http.StatusOK {
			_, err := fw.w.Write(flvHeader[:])
			if err != nil {
				return 0, err
			}
		}
	}
	return fw.w.Write(b)
}

func (fw *flvWriter) WriteHeader(s int) {
	if s == http.StatusPartialContent {
		s = http.StatusOK
		fw.w.Header().Del("Content-Range")  // don't need that
		fw.w.Header().Del("Content-Length") // we can recalculate it
	}
	fw.w.WriteHeader(s)
	fw.status = s
}
