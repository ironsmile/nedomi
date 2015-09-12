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
	h.Logger.Debugf("[%p] Cacheable access: %s", h.req, h.req.RequestURI)

	obj, err := h.Cache.Storage.GetMetadata(h.objID)
	if os.IsNotExist(err) {
		h.Logger.Debugf("[%p] No metadata on storage, proxying...", h.req)
		h.carbonCopyProxy()
	} else if err != nil {
		h.Logger.Logf("[%p] Storage error when reading metadata: %s", h.req, err)
		h.Cache.Storage.Discard(h.objID)
		h.carbonCopyProxy()
	} else if !utils.IsMetadataFresh(obj) {
		h.Logger.Logf("[%p] Metadata is stale, refreshing...", h.req)
		//!TODO: optimize, do only a head request when the metadata is stale?
		h.Cache.Storage.Discard(h.objID)
		h.carbonCopyProxy()
	} else {
		h.obj = obj

		//!TODO: rewrite date header?
		if h.req.Header.Get("Range") != "" {
			h.knownRanged()
		} else {
			h.knownFull()
		}
	}
}

func (h *reqHandler) carbonCopyProxy() {
	//!TODO: consult the cache algorithm whether to save the metadata
	hook := h.getResponseHook(true)
	flexibleResp := utils.NewFlexibleResponseWriter(hook)

	h.Upstream.ServeHTTP(flexibleResp, h.getNormalizedRequest())
	flexibleResp.BodyWriter.Close()
}

func (h *reqHandler) knownRanged() {
	h.Logger.Logf("[%p] Metadata found, downloading only missing parts...", h.req)

	if h.obj.Size == nil {
		// We do not know the size of the object, so we cannot satisfy range requests
		h.carbonCopyProxy()
		return
	}
	size := int64(*h.obj.Size)

	ranges, err := parseRange(h.req.Header.Get("Range"), size)
	if err != nil {
		err := http.StatusRequestedRangeNotSatisfiable
		http.Error(h.resp, http.StatusText(err), err)
		return
	}

	if len(ranges) != 1 {
		//!TODO: implement support for multiple ranges
		// We do not support multiple ranges but maybe the upstream does
		h.carbonCopyProxy()
		return
	}
	reqRange := ranges[0]

	for k := range h.obj.Headers {
		h.resp.Header()[k] = h.obj.Headers[k]
	}
	h.resp.Header().Set("Content-Range", reqRange.contentRange(size))
	h.resp.Header().Set("Content-Length", strconv.FormatInt(reqRange.length, 10))
	h.resp.WriteHeader(http.StatusPartialContent)

	end := uint64(ranges[0].start + ranges[0].length - 1)
	reader := h.getPartReader(uint64(ranges[0].start), &end)
	if copied, err := io.Copy(h.resp, reader); err != nil {
		h.Logger.Errorf("[%p] Error copying response: %s. Copied %d", h.req, err, copied)
	}
	reader.Close()
}

func (h *reqHandler) knownFull() {
	for k := range h.obj.Headers {
		h.resp.Header()[k] = h.obj.Headers[k]
	}
	h.resp.WriteHeader(h.obj.Code)

	reader := h.getPartReader(0, h.obj.Size)
	if copied, err := io.Copy(h.resp, reader); err != nil {
		h.Logger.Errorf("[%p] Error copying response: %s. Copied %d", h.req, err, copied)
	}
	reader.Close()
}
