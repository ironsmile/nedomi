package cache

import (
	"io"
	"net/http"
	"os"
	"strconv"

	"golang.org/x/net/context"

	"github.com/ironsmile/nedomi/types"
	"github.com/ironsmile/nedomi/utils"
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
}

// handle tries to respond to client request by loading metadata and file parts
// from the cache. If there are missing parts, they are retrieved from the upstream.
func (h *reqHandler) handle() {
	h.Logger.Debugf("[%p] Caching proxy access: %s", h.req, h.req.RequestURI)

	rng := h.req.Header.Get("Range")
	obj, err := h.Cache.Storage.GetMetadata(h.objID)
	if os.IsNotExist(err) {
		h.Logger.Debugf("[%p] No metadata on storage, proxying...", h.req)
		h.carbonCopyProxy()
	} else if err != nil {
		h.Logger.Errorf("[%p] Storage error when reading metadata: %s", h.req, err)
		h.Cache.Storage.Discard(h.objID)
		h.carbonCopyProxy()
	} else if !utils.IsMetadataFresh(obj) {
		h.Logger.Debugf("[%p] Metadata is stale, proxying...", h.req)
		//!TODO: optimize, do only a head request when the metadata is stale?
		h.Cache.Storage.Discard(h.objID)
		h.carbonCopyProxy()
	} else if !utils.CacheSatisfiesRequest(obj, h.req) {
		h.Logger.Debugf("[%p] Client does not want cached response or the cache does not satisfy the request, proxying...", h.req)
		h.carbonCopyProxy()
	} else {
		h.obj = obj
		//!TODO: rewrite date header?

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
	//!TODO: consult the cache algorithm whether to save the metadata
	hook := h.getResponseHook()
	flexibleResp := utils.NewFlexibleResponseWriter(hook)

	h.Upstream.ServeHTTP(flexibleResp, h.getNormalizedRequest())
	flexibleResp.BodyWriter.Close()
}

func (h *reqHandler) knownRanged() {
	ranges, err := parseReqRange(h.req.Header.Get("Range"), h.obj.Size)
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

	utils.CopyHeadersWithout(h.obj.Headers, h.resp.Header())
	h.resp.Header().Set("Content-Range", reqRange.contentRange(h.obj.Size))
	h.resp.Header().Set("Content-Length", strconv.FormatUint(reqRange.length, 10))
	h.resp.WriteHeader(http.StatusPartialContent)

	end := ranges[0].start + ranges[0].length - 1
	reader := h.getSmartReader(ranges[0].start, end)
	if copied, err := io.Copy(h.resp, reader); err != nil {
		h.Logger.Errorf("[%p] Error copying response: %s. Copied %d out of %d bytes", h.req, err, copied, reqRange.length)
	} else if uint64(copied) != reqRange.length {
		h.Logger.Errorf("[%p] Error copying response. Expected to copy %d bytes, copied %d", h.req, reqRange.length, copied)
	}
	reader.Close()
}

func (h *reqHandler) knownFull() {
	utils.CopyHeadersWithout(h.obj.Headers, h.resp.Header())
	h.resp.Header().Set("Content-Length", strconv.FormatUint(h.obj.Size, 10))
	h.resp.WriteHeader(h.obj.Code)

	reader := h.getSmartReader(0, h.obj.Size)
	if copied, err := io.Copy(h.resp, reader); err != nil {
		h.Logger.Errorf("[%p] Error copying response: %s. Copied %d out of %d bytes", h.req, err, copied, h.obj.Size)
	} else if uint64(copied) != h.obj.Size {
		h.Logger.Errorf("[%p] Error copying response. Expected to copy %d bytes, copied %d", h.req, h.obj.Size, copied)
	}
	reader.Close()
}
