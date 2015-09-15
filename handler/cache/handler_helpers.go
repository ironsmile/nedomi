package cache

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/ironsmile/nedomi/storage"
	"github.com/ironsmile/nedomi/types"
	"github.com/ironsmile/nedomi/utils"
)

func copyHeadersWithout(from, to http.Header, exceptions ...string) {
	for k := range from {
		shouldCopy := true
		for _, e := range exceptions {
			if e == k {
				shouldCopy = false
				break
			}
		}
		if shouldCopy {
			to[k] = from[k]
		}
	}
}

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

	copyHeadersWithout(h.req.Header, result.Header, "Accept-Encoding")

	//!TODO: fix requested range to be divisible by the storage partSize

	return result
}

func (h *reqHandler) getDimensions(code int, headers http.Header) (*httpContentRange, error) {
	rangeStr := headers.Get("Content-Range")
	lengthStr := headers.Get("Content-Length")
	if code == http.StatusPartialContent {
		if rangeStr != "" {
			return parseRespContentRange(rangeStr)
		}
		return nil, errors.New("No Content-Range header")
	} else if code == http.StatusOK {
		if lengthStr != "" {
			size, err := strconv.ParseUint(lengthStr, 10, 64)
			if err != nil {
				return nil, err
			}
			return &httpContentRange{start: 0, length: size, objSize: size}, nil
		}
		return nil, errors.New("No Content-Length header")
	}
	return nil, errors.New("Invalid HTTP status or no object length data in the headers")
}

func (h *reqHandler) getResponseHook() func(*utils.FlexibleResponseWriter) {

	return func(rw *utils.FlexibleResponseWriter) {
		h.Logger.Debugf("[%p] Received headers for %s, sending them to client...", h.req, h.req.URL)
		copyHeadersWithout(rw.Headers, h.resp.Header())
		h.resp.WriteHeader(rw.Code)

		//!TODO: handle duration
		isCacheable, _ := utils.IsResponseCacheable(rw.Code, rw.Headers)
		dims, err := h.getDimensions(rw.Code, rw.Headers)
		if !isCacheable || err != nil {
			h.Logger.Debugf("[%p] Response is non-cacheable (%s) :(", h.req, err)
			rw.BodyWriter = storage.NopCloser(h.resp)
			return
		}

		h.Logger.Debugf("[%p] Response is cacheable! Caching metadata and parts...", h.req)

		obj := &types.ObjectMetadata{
			ID:                h.objID,
			ResponseTimestamp: time.Now().Unix(),
			Code:              rw.Code,
			Size:              dims.objSize,
			Headers:           make(http.Header),
		}
		copyHeadersWithout(rw.Headers, obj.Headers, "Transfer-Encoding", "Content-Length", "Content-Range")

		//!TODO: optimize this, save the metadata only when it's newer
		//!TODO: also, error if we already have fresh metadata but the received metadata is different
		if err := h.Cache.Storage.SaveMetadata(obj); err != nil {
			h.Logger.Errorf("Could not save metadata for %s: %s", obj.ID, err)
		}

		//!TODO: handle range requests
		rw.BodyWriter = storage.MultiWriteCloser(
			storage.NopCloser(h.resp),
			storage.PartWriter(h.Cache.Storage, h.objID, dims.start, dims.length),
		)
	}
}

func (h *reqHandler) getUpstreamReader(start, end uint64) io.ReadCloser {
	subh := *h
	subh.req = subh.getNormalizedRequest()
	subh.req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", start, end))

	h.Logger.Debugf("[%p] Making upstream request for %s, bytes [%d-%d]...",
		subh.req, subh.req.URL, start, end)

	//!TODO: optimize requests for the same pieces? if possible, make only 1 request to the upstream for the same part

	r, w := io.Pipe()
	subh.resp = utils.NewFlexibleResponseWriter(func(rw *utils.FlexibleResponseWriter) {
		h.Logger.Debugf("[%p] Received response with status %d", subh.req, rw.Code)
		if rw.Code == http.StatusPartialContent {
			rw.BodyWriter = w
		} else if rw.Code == http.StatusOK {
			//!TODO: handle this, use skipWriter or something like that
			w.CloseWithError(fmt.Errorf("NOT IMPLEMENTED"))
		} else {
			w.CloseWithError(fmt.Errorf("Upstream responded with status %d", rw.Code))
		}
	})
	go subh.carbonCopyProxy()
	return r
}

func (h *reqHandler) getSmartReader(start, end uint64) io.ReadCloser {

	partSize := h.Cache.Storage.PartSize()
	localCount := 0
	indexes := utils.BreakInIndexes(h.objID, start, end, partSize)
	lastPresentIndex := -1
	readers := []io.ReadCloser{}

	h.Logger.Debugf("[%p] Trying to load all possible parts of %s from storage...", h.req, h.objID)
	for i, idx := range indexes {
		r, err := h.Cache.Storage.GetPart(idx)
		if err != nil {
			if !os.IsNotExist(err) {
				h.Logger.Errorf("[%p] Unexpected error while trying to load %s from storage: %s", h.req, idx, err)
			}
			continue
		}

		if lastPresentIndex != i-1 {
			fromPart := uint64(lastPresentIndex + 1)
			toPart := uint64(i - 1)
			h.Logger.Debugf("[%p] Getting parts [%d-%d] from upstream!", h.req, fromPart, toPart)
			readers = append(readers, h.getUpstreamReader(fromPart*partSize, (toPart+1)*partSize-1))
		}
		h.Logger.Debugf("[%p] Loaded part %s from storage!", h.req, idx)
		localCount++
		readers = append(readers, r)
		lastPresentIndex = i
	}

	// work in start and end
	var startOffset, endLimit = start % partSize, end%partSize + 1
	readers[0] = storage.SkipReadCloser(readers[0], int(startOffset))
	readers[len(readers)-1] = storage.LimitReadCloser(readers[len(readers)-1], int(endLimit))

	h.Logger.Debugf("[%p] Return smart reader for %s with %d out of %d parts from storage!",
		h.req, h.objID, localCount, len(indexes))
	return storage.MultiReadCloser(readers...)
}
