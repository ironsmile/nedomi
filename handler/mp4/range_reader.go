package mp4

import (
	"fmt"
	"io"
	"net/http"

	"golang.org/x/net/context"

	"github.com/ironsmile/nedomi/contexts"
	"github.com/ironsmile/nedomi/types"
	"github.com/ironsmile/nedomi/utils/httputils"
)

type rangeReader struct {
	reqID    types.RequestID
	ctx      context.Context
	req      *http.Request
	location *types.Location
	next     types.RequestHandler
	callback func(frw *httputils.FlexibleResponseWriter) bool
}

func (rr *rangeReader) Range(start, length uint64) io.ReadCloser {
	newreq := copyRequest(rr.req)
	newreq.Header.Set("Range", httputils.Range{Start: start, Length: length}.Range())
	var newCtx, newID = contexts.AppendToRequestID(rr.ctx, []byte(fmt.Sprintf("mp4=%d+%d", start, length)))
	var in, out = io.Pipe()
	flexible := httputils.NewFlexibleResponseWriter(func(frw *httputils.FlexibleResponseWriter) {
		if frw.Code != http.StatusPartialContent || !rr.callback(frw) {
			_ = out.CloseWithError(errUnsatisfactoryResponse)
		}
		frw.BodyWriter = out
	})
	go func() {
		defer func() {
			if err := out.Close(); err != nil {
				rr.location.Logger.Errorf("[%s]: error on closing rangeReaders output: %s",
					newID, err)
			}
		}()
		rr.next.ServeHTTP(newCtx, flexible, newreq)
	}()

	return in
}

func (rr *rangeReader) RangeRead(start, length uint64) (io.ReadCloser, error) {
	return rr.Range(start, length), nil
}
