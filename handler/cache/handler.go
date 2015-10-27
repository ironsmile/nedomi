package cache

import (
	"net/http"
	"os"
	"strconv"
	"time"

	"golang.org/x/net/context"

	"github.com/ironsmile/nedomi/types"
	"github.com/ironsmile/nedomi/utils"
	"github.com/ironsmile/nedomi/utils/cacheutils"
	"github.com/ironsmile/nedomi/utils/httputils"
)

// reqHandler handles an individual request (so we don't have to pass a lot of
// parameters and state between the different functions)
type reqHandler struct {
	*CachingProxy
	ctx   context.Context
	req   *http.Request
	resp  responseWriteCloser
	objID *types.ObjectID
	obj   *types.ObjectMetadata
}

// handle tries to respond to client request by loading metadata and file parts
// from the cache. If there are missing parts, they are retrieved from the upstream.
func (h *reqHandler) handle() {
	h.Logger.Debugf("[%p] Caching proxy access: %s %s", h.req, h.req.Method, h.req.RequestURI)

	rng := h.req.Header.Get("Range")
	obj, err := h.Cache.Storage.GetMetadata(h.objID)
	if os.IsNotExist(err) {
		h.Logger.Debugf("[%p] No metadata on storage, proxying...", h.req)
		h.carbonCopyProxy()
	} else if err != nil {
		h.Logger.Errorf("[%p] Storage error when reading metadata: %s", h.req, err)
		if isTooManyFiles(err) {
			http.Error(
				h.resp,
				http.StatusText(http.StatusInternalServerError),
				http.StatusInternalServerError)
			return
		}
		if discardErr := h.Cache.Storage.Discard(h.objID); discardErr != nil {
			h.Logger.Errorf("[%p] Storage error when discarding of object's data: %s", h.req, discardErr)
		}
		h.carbonCopyProxy()
	} else if !utils.IsMetadataFresh(obj) {
		h.Logger.Debugf("[%p] Metadata is stale, proxying...", h.req)
		//!TODO: optimize, do only a head request when the metadata is stale?
		if discardErr := h.Cache.Storage.Discard(h.objID); discardErr != nil {
			h.Logger.Errorf("[%p] Storage error when discarding of object's data: %s", h.req, discardErr)
		}
		h.carbonCopyProxy()
	} else if !cacheutils.CacheSatisfiesRequest(obj, h.req) {
		h.Logger.Debugf("[%p] Client does not want cached response or the cache does not satisfy the request, proxying...", h.req)
		h.carbonCopyProxy()
	} else {
		h.obj = obj
		//!TODO: advertise that we support ranges - send "Accept-Ranges: bytes"?

		//!TODO: evaluate conditional requests: https://tools.ietf.org/html/rfc7232
		//!TODO: Also, handle this from RFC7233:
		// "The Range header field is evaluated after evaluating the precondition
		// header fields defined in [RFC7232], and only if the result in absence
		// of the Range header field would be a 200 (OK) response.  In other
		// words, Range is ignored when a conditional GET would result in a 304
		// (Not Modified) response."

		if rng != "" {
			h.Logger.Debugf("[%p] Serving range '%s', preferably from cache...", h.req, rng)
			h.knownRanged()
		} else {
			h.Logger.Debugf("[%p] Serving full object, preferably from cache...", h.req)
			h.knownFull()
		}
	}
}

func (h *reqHandler) carbonCopyProxy() {
	flexibleResp := httputils.NewFlexibleResponseWriter(h.getResponseHook())
	defer func() {
		if flexibleResp.BodyWriter != nil {
			if err := flexibleResp.BodyWriter.Close(); err != nil {
				h.Logger.Errorf("[%p] Error while closing flexibleResponse: %s", h.req, err)
			}
		}
		//!TODO: cache small upstream responses that we did not cache because
		// there was no Content-Length header in the upstream response but it
		// was otherwise cacheable? Examples are folder listings for apache and
		// ngingx: `curl -i http://mirror.rackspace.com/` or `curl -i
		// https://mirrors.uni-plovdiv.net/`

	}()

	h.next.RequestHandle(h.ctx, flexibleResp, h.getNormalizedRequest())
}

func (h *reqHandler) knownRanged() {
	ranges, err := httputils.ParseRequestRange(h.req.Header.Get("Range"), h.obj.Size)
	if err != nil {
		err := http.StatusRequestedRangeNotSatisfiable
		http.Error(h.resp, http.StatusText(err), err)
		return
	}

	if len(ranges) != 1 {
		// We do not support multiple ranges but maybe the upstream does
		//!TODO: implement support for multiple ranges
		h.carbonCopyProxy()
		return
	}
	reqRange := ranges[0]

	httputils.CopyHeaders(h.obj.Headers, h.resp.Header())
	h.resp.Header().Set("Content-Range", reqRange.ContentRange(h.obj.Size))
	h.resp.Header().Set("Content-Length", strconv.FormatUint(reqRange.Length, 10))
	h.rewriteTimeBasedHeaders()
	h.resp.WriteHeader(http.StatusPartialContent)
	if h.req.Method == "HEAD" {
		return
	}

	h.lazilyRespond(ranges[0].Start, ranges[0].Start+ranges[0].Length-1)
}

func (h *reqHandler) knownFull() {
	httputils.CopyHeaders(h.obj.Headers, h.resp.Header())
	h.resp.Header().Set("Content-Length", strconv.FormatUint(h.obj.Size, 10))
	h.rewriteTimeBasedHeaders()
	h.resp.WriteHeader(h.obj.Code)
	if h.req.Method == "HEAD" {
		return
	}

	h.lazilyRespond(0, h.obj.Size-1)
}

func (h *reqHandler) rewriteTimeBasedHeaders() {
	var now = time.Now()
	h.resp.Header().Set("Date", now.Format(http.TimeFormat))
	h.resp.Header().Set("Expires", time.Unix(h.obj.ExpiresAt, 0).Format(http.TimeFormat))
	h.resp.Header().Set("Age", strconv.FormatInt(now.Unix()-h.obj.ResponseTimestamp, 10))
	h.resp.Header().Set("Cache-Control", "max-age="+strconv.FormatInt(h.obj.ExpiresAt-now.Unix(), 10))
}
