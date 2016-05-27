package cache

import (
	"net/http"
	"os"
	"strconv"
	"time"

	"golang.org/x/net/context"

	"github.com/ironsmile/nedomi/contexts"
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
	resp  http.ResponseWriter
	objID *types.ObjectID
	obj   *types.ObjectMetadata
	reqID types.RequestID
}

// handle tries to respond to client request by loading metadata and file parts
// from the cache. If there are missing parts, they are retrieved from the upstream.
func (h *reqHandler) handle() {
	h.objID = h.NewObjectIDForURL(h.req.URL)
	h.reqID, _ = contexts.GetRequestID(h.ctx)
	h.Logger.Debugf("[%s] Caching proxy access: %s %s", h.reqID, h.req.Method, h.req.RequestURI)

	rng := h.req.Header.Get("Range")
	obj, err := h.Cache.Storage.GetMetadata(h.objID)
	if os.IsNotExist(err) {
		h.Logger.Debugf("[%s] No metadata on storage, proxying...", h.reqID)
		h.carbonCopyProxy()
	} else if err != nil {
		h.Logger.Errorf("[%s] Storage error when reading metadata: %s", h.reqID, err)
		if isTooManyFiles(err) {
			http.Error(
				h.resp,
				http.StatusText(http.StatusInternalServerError),
				http.StatusInternalServerError)
			return
		}
		if discardErr := h.Cache.Storage.Discard(h.objID); discardErr != nil {
			h.Logger.Errorf("[%s] Storage error when discarding of object's data: %s",
				h.reqID, discardErr)
		}
		h.carbonCopyProxy()
	} else if !utils.IsMetadataFresh(obj) {
		h.Logger.Debugf("[%s] Metadata is stale, proxying...", h.reqID)
		//!TODO: optimize, do only a head request when the metadata is stale?
		if discardErr := h.Cache.Storage.Discard(h.objID); discardErr != nil {
			h.Logger.Errorf("[%s] Storage error when discarding of object's data: %s",
				h.reqID, discardErr)
		}
		h.carbonCopyProxy()
	} else if !cacheutils.CacheSatisfiesRequest(obj, h.req) {
		h.Logger.Debugf("[%s] Client does not want cached response or the cache does not"+
			"satisfy the request, proxying...", h.reqID)
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
			h.Logger.Debugf("[%s] Serving range '%s', preferably from cache...",
				h.reqID, rng)
			h.knownRanged()
		} else {
			h.Logger.Debugf("[%s] Serving full object, preferably from cache...",
				h.reqID)
			h.knownFull()
		}
	}
}

func (h *reqHandler) carbonCopyProxy() {
	flexibleResp := httputils.NewFlexibleResponseWriter(h.getResponseHook())
	defer func() {
		if flexibleResp.BodyWriter != nil {
			if err := flexibleResp.BodyWriter.Close(); err != nil {
				if isPartWriterShorWrite(err) {
					h.Logger.Debugf("[%s] Error while closing flexibleResponse: %s", h.reqID, err)
				} else {
					h.Logger.Errorf("[%s] Error while closing flexibleResponse: %s", h.reqID, err)
				}
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

	var responseSize = h.obj.Size
	if responseSize != 0 {
		responseSize--
	}
	h.lazilyRespond(0, responseSize)
}

func (h *reqHandler) rewriteTimeBasedHeaders() {
	var nowUnix = time.Now().Unix()
	h.resp.Header().Set("Expires", time.Unix(h.obj.ExpiresAt, 0).Format(http.TimeFormat))
	h.resp.Header().Set("Age", strconv.FormatInt(nowUnix-h.obj.ResponseTimestamp, 10))
	h.resp.Header().Set("Cache-Control", "max-age="+strconv.FormatInt(h.obj.ExpiresAt-nowUnix, 10))
}

func isPartWriterShorWrite(err error) bool {
	if o, ok := err.(interface {
		Original() error
	}); ok {
		return isPartWriterShorWrite(o.Original())
	}
	if ce, ok := err.(*utils.CompositeError); ok {
		for _, err = range *ce {
			if isPartWriterShorWrite(err) {
				return true
			}
		}
		return false
	}
	_, ok := err.(*partWriterShortWrite)
	return ok
}
