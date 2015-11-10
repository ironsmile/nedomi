package cache

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"syscall"
	"time"

	"github.com/ironsmile/nedomi/storage"
	"github.com/ironsmile/nedomi/types"
	"github.com/ironsmile/nedomi/utils"
	"github.com/ironsmile/nedomi/utils/cacheutils"
	"github.com/ironsmile/nedomi/utils/httputils"
)

// Hop-by-hop headers. These are removed when sent to the client.
var hopHeaders = httputils.GetHopByHopHeaders()

var metadataHeadersToFilter = append(hopHeaders, "Content-Length", "Content-Range", "Expires", "Age", "Cache-Control")

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

	httputils.CopyHeadersWithout(h.req.Header, result.Header, "Accept-Encoding")

	//!TODO: fix requested range to be divisible by the storage partSize

	return result
}

func (h *reqHandler) getResponseHook() func(*httputils.FlexibleResponseWriter) {

	return func(rw *httputils.FlexibleResponseWriter) {
		h.Logger.Debugf("[%p] Received headers for %s, sending them to client...", h.req, h.req.URL)
		httputils.CopyHeadersWithout(rw.Headers, h.resp.Header(), hopHeaders...)
		h.resp.WriteHeader(rw.Code)

		isCacheable := cacheutils.IsResponseCacheable(rw.Code, rw.Headers)
		if !isCacheable {
			h.Logger.Debugf("[%p] Response is non-cacheable", h.req)
			rw.BodyWriter = utils.AddCloser(h.resp)
			return
		}

		expiresIn := cacheutils.ResponseExpiresIn(rw.Headers, h.CacheDefaultDuration)
		if expiresIn <= 0 {
			h.Logger.Debugf("[%p] Response expires in the past: %s", h.req, expiresIn)
			rw.BodyWriter = utils.AddCloser(h.resp)
			return
		}

		responseRange, err := httputils.GetResponseRange(rw.Code, rw.Headers)
		if err != nil {
			h.Logger.Debugf("[%p] Was not able to get response range (%s)", h.req, err)
			rw.BodyWriter = utils.AddCloser(h.resp)
			return
		}

		h.Logger.Debugf("[%p] Response is cacheable! Caching metadata and parts...", h.req)

		code := rw.Code
		if code == http.StatusPartialContent {
			// 206 is returned only if the server would have returned 200 with a normal request
			code = http.StatusOK
		}

		//!TODO: maybe call cached time.Now. See the comment in utils.IsMetadataFresh
		now := time.Now()

		obj := &types.ObjectMetadata{
			ID:                h.objID,
			ResponseTimestamp: now.Unix(),
			Code:              code,
			Size:              responseRange.ObjSize,
			Headers:           make(http.Header),
			ExpiresAt:         now.Add(expiresIn).Unix(),
		}
		httputils.CopyHeadersWithout(rw.Headers, obj.Headers, metadataHeadersToFilter...)
		if obj.Headers.Get("Date") == "" { // maybe the server does not return date, we should set it then
			obj.Headers.Set("Date", now.Format(http.TimeFormat))
		}

		//!TODO: consult the cache algorithm whether to save the metadata
		//!TODO: optimize this, save the metadata only when it's newer
		//!TODO: also, error if we already have fresh metadata but the received metadata is different
		if err := h.Cache.Storage.SaveMetadata(obj); err != nil {
			h.Logger.Errorf("[%p] Could not save metadata for %s: %s", h.req, obj.ID, err)
			rw.BodyWriter = utils.AddCloser(h.resp)
			return
		}

		if h.req.Method == "HEAD" {
			rw.BodyWriter = utils.AddCloser(h.resp)
			return
		}

		rw.BodyWriter = utils.MultiWriteCloser(
			utils.AddCloser(h.resp),
			PartWriter(h.Cache, h.objID, *responseRange),
		)

		h.Logger.Debugf("[%p] Setting the cached data to expire in %s", h.req, expiresIn)
		h.Cache.Scheduler.AddEvent(
			h.objID.Hash(),
			storage.GetExpirationHandler(h.Cache, h.Logger, h.objID),
			expiresIn,
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
	subh.resp = httputils.NewFlexibleResponseWriter(func(rw *httputils.FlexibleResponseWriter) {
		respRng, err := httputils.GetResponseRange(rw.Code, rw.Headers)
		if err != nil {
			h.Logger.Debugf("[%p] Could not parse the content-range for the partial upstream request: %s", subh.req, err)
			_ = w.CloseWithError(err)
		}
		h.Logger.Debugf("[%p] Received response with status %d and range %v", subh.req, rw.Code, respRng)
		if rw.Code == http.StatusPartialContent {
			//!TODO: check whether the returned range corresponds to the requested range
			rw.BodyWriter = w
		} else if rw.Code == http.StatusOK {
			//!TODO: handle this, use skipWriter or something like that
			_ = w.CloseWithError(fmt.Errorf("NOT IMPLEMENTED"))
		} else {
			_ = w.CloseWithError(fmt.Errorf("Upstream responded with status %d", rw.Code))
		}
	})
	go subh.carbonCopyProxy()
	return newWholeChunkReadCloser(r, h.Cache.PartSize.Bytes())
}

// if error is returned - it is 'too many open files'
func (h *reqHandler) getPartFromStorage(idx *types.ObjectIndex) (io.ReadCloser, error) {
	cached := h.Cache.Algorithm.Lookup(idx)
	r, err := h.Cache.Storage.GetPart(idx)
	if err == nil {
		h.Cache.Algorithm.PromoteObject(idx)
		return r, nil
	}
	if !os.IsNotExist(err) {
		if isTooManyFiles(err) {
			return nil, err
		}
		h.Logger.Errorf("[%p] Unexpected error while trying to load %s from storage: %s", h.req, idx, err)
	} else if cached {
		h.Logger.Debugf("[%p] Cache.Algorithm said a part %s is cached but Storage couldn't find it", h.req, idx)
	}
	return nil, nil
}

func (h *reqHandler) getContents(indexes []*types.ObjectIndex, from int) (io.ReadCloser, int, error) {
	r, err := h.getPartFromStorage(indexes[from])
	if r != nil {
		return r, 1, nil
	} else if err != nil {
		return nil, 0, err
	}

	partSize := h.Cache.Storage.PartSize()
	fromByte := uint64(indexes[from].Part) * partSize
	for to := from + 1; to < len(indexes); to++ {
		r, _ = h.getPartFromStorage(indexes[to])
		if r != nil {
			toByte := umin(h.obj.Size, uint64(indexes[to].Part)*partSize-1)
			return utils.MultiReadCloser(h.getUpstreamReader(fromByte, toByte), r), to - from + 1, nil
		}
	}

	toByte := umin(h.obj.Size, uint64(indexes[len(indexes)-1].Part+1)*partSize-1)
	return h.getUpstreamReader(fromByte, toByte), len(indexes) - from, nil
}

func (h *reqHandler) lazilyRespond(start, end uint64) {
	partSize := h.Cache.Storage.PartSize()
	indexes := utils.BreakInIndexes(h.objID, start, end, partSize)
	startOffset := start % partSize
	var shouldReturn = false

	for i := 0; i < len(indexes); {
		contents, partsCount, err := h.getContents(indexes, i)
		if err != nil {
			h.Logger.Errorf("[%p] Unexpected error while trying to load %s from storage: %s", h.req, indexes[i], err)
			return
		}
		if i == 0 && startOffset > 0 {
			contents, err = utils.SkipReadCloser(contents, int64(startOffset))
			if err != nil {
				h.Logger.Errorf("[%p] Unexpected error while trying to skip %d from %s : %s", h.req, startOffset, indexes[i], err)
				return
			}
		}
		if i+partsCount == len(indexes) {
			endLimit := uint64(partsCount-1)*partSize + end%partSize + 1
			if i == 0 {
				endLimit -= startOffset
			}
			contents = utils.LimitReadCloser(contents, int64(endLimit)) //!TODO: fix int conversion
		}

		if copied, err := io.Copy(h.resp, contents); err != nil {
			h.Logger.Logf("[%p] Error sending contents after %d bytes of %s, parts[%d-%d]: %s",
				h.req, copied, h.objID, i, i+partsCount-1, err)

			shouldReturn = true
		}
		//!TODO: compare the copied length with the expected

		if err := contents.Close(); err != nil {
			h.Logger.Errorf("[%p] Unexpected error while closing content reader for %s, parts[%d-%d]: %s",
				h.req, h.objID, i, i+partsCount-1, err)
		}

		if shouldReturn {
			return
		}

		i += partsCount
	}
}

func isTooManyFiles(err error) bool {
	if pathError, ok := err.(*os.PathError); ok {
		return pathError.Err == syscall.EMFILE
	}
	return err == syscall.EMFILE
}
