// Package headers could be used to rewrite headers of requests and responses.
package headers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"golang.org/x/net/context"

	"github.com/ironsmile/nedomi/config"
	"github.com/ironsmile/nedomi/types"
	"github.com/ironsmile/nedomi/utils"
	"github.com/ironsmile/nedomi/utils/httputils"
)

// Headers rewrites headers
type Headers struct {
	next     types.RequestHandler
	request  headersRewrite
	response headersRewrite
}

type headersConfig struct {
	Request  config.HeadersRewrite `json:"request"`
	Response config.HeadersRewrite `json:"response"`
}

var emptyHeadersRewrite = headersRewrite{}

// RequestHandle rewrites the headers of the given request
func (h *Headers) RequestHandle(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	if !h.request.isEmpty() {
		h.request.rewrite(r.Header)
	}
	if !h.response.isEmpty() {
		w = h.wrapResponseWriter(w)
	}
	h.next.RequestHandle(ctx, w, r)
}

// New creates and returns a ready to used ServerStatusHandler.
func New(cfg *config.Handler, l *types.Location, next types.RequestHandler) (*Headers, error) {
	var hr headersConfig
	if len(cfg.Settings) != 0 {
		if err := json.Unmarshal(cfg.Settings, &hr); err != nil {
			return nil, err
		}
	}
	return NewHeaders(next, hr.Request, hr.Response)
}

// NewHeaders is a more convinient constructor
func NewHeaders(next types.RequestHandler, request, response config.HeadersRewrite) (*Headers, error) {
	if next == nil {
		return nil, fmt.Errorf("headers handler requires next handler")
	}
	return &Headers{
		next:     next,
		request:  headersRewrite(request),
		response: headersRewrite(response),
	}, nil
}

func (h *Headers) wrapResponseWriter(w http.ResponseWriter) http.ResponseWriter {
	var newW = httputils.NewFlexibleResponseWriter(func(frw *httputils.FlexibleResponseWriter) {
		httputils.CopyHeaders(frw.Header(), w.Header())
		h.response.rewrite(w.Header())
		frw.BodyWriter = utils.NopCloser(w)
		w.WriteHeader(frw.Code)
	})
	httputils.CopyHeaders(w.Header(), newW.Header())
	return newW
}
