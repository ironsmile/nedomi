package cache

import (
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/ironsmile/nedomi/storage"
	"github.com/ironsmile/nedomi/types"
	"github.com/ironsmile/nedomi/utils"
)

// Returns a new HTTP 1.1 request that has no body. It also clears headers like
// accept-encoding and rearranges the requested ranges so they match part
func (h *reqHandler) getNormalizedRequest() *http.Request {
	url := *h.req.URL
	result := &http.Request{
		Method:     h.req.Method,
		URL:        &url,
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		Header:     make(http.Header),
		Host:       h.req.URL.Host,
	}

	for k := range h.req.Header {
		if k != "Accept-Encoding" {
			result.Header[k] = h.req.Header[k]
		}
	}

	//!TODO: fix requested range

	return result
}

func (h *reqHandler) getResponseHook(saveMetadata bool) func(*utils.FlexibleResponseWriter) {

	return func(rw *utils.FlexibleResponseWriter) {
		h.Logger.Logf("[%p] Received headers for %s, sending them to client...", h.req, h.req.URL)
		for k := range rw.Headers {
			h.resp.Header()[k] = rw.Headers[k]
		}
		h.resp.WriteHeader(rw.Code)

		//!TODO: handle duration
		isCacheable, _ := utils.IsResponseCacheable(rw.Code, rw.Headers)
		if !isCacheable || rw.Header().Get("Content-Range") != "" { //!TODO: remove content-range check
			h.Logger.Logf("[%p] Response is non-cacheable :(", h.req)
			rw.BodyWriter = storage.NopCloser(h.resp)
			return
		}

		h.Logger.Logf("[%p] Response is cacheable! Caching parts...", h.req)
		if saveMetadata {
			//!TODO: fix, this is wrong for range requests

			obj := &types.ObjectMetadata{
				ID:                h.objID,
				ResponseTimestamp: time.Now().Unix(),
				Code:              rw.Code,
				Headers:           rw.Headers, //!TODO: remove Transfer-Encoding and Content-Range headers
			}

			cl := h.resp.Header().Get("Content-Length")
			contentLength, err := strconv.ParseUint(cl, 10, 64)
			if err != nil {
				obj.Size = &contentLength
			}

			if err := h.Cache.Storage.SaveMetadata(obj); err != nil {
				h.Logger.Errorf("Could not save metadata for %s: %s", obj.ID, err)
			}
		}

		//!TODO: handle missing content length, partial requests (range headers), etc.
		rw.BodyWriter = storage.MultiWriteCloser(
			storage.NopCloser(h.resp),
			storage.PartWriter(h.Cache.Storage, h.objID, 0),
		)
	}
}

func (h *reqHandler) getPartReader(start uint64, end *uint64) io.ReadCloser {
	//!TODO: handle ranges correctly, handle *unknown* object size correctly
	readers := []io.ReadCloser{}
	var part uint32
	for {
		idx := &types.ObjectIndex{ObjID: h.objID, Part: part}
		h.Logger.Debugf("[%p] Trying to load part %s from storage...", h.req, idx)
		r, err := h.Cache.Storage.GetPart(idx)
		if err != nil {
			break //TODO: fix, this is wrong
		}
		h.Logger.Debugf("[%p] Loaded part %s from storage!", h.req, idx)
		readers = append(readers, r)
		part++
	}
	h.Logger.Debugf("[%p] Return reader with %d parts of %s from storage!", h.req, len(readers), h.objID)
	return storage.MultiReadCloser(readers...)
}
