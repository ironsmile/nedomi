package cache

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"strconv"
	"syscall"
	"time"

	"github.com/ironsmile/nedomi/contexts"
	"github.com/ironsmile/nedomi/storage"
	"github.com/ironsmile/nedomi/types"
	"github.com/ironsmile/nedomi/utils"
	"github.com/ironsmile/nedomi/utils/cacheutils"
	"github.com/ironsmile/nedomi/utils/httputils"
)

// Hop-by-hop headers. These are removed when sent to the client.
var hopHeaders = httputils.GetHopByHopHeaders()

var metadataHeadersToFilter = append(hopHeaders,
	"Content-Length", "Content-Range", "Expires", "Age", "Cache-Control")

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
		Host:       h.req.Host,
	}

	httputils.CopyHeadersWithout(h.req.Header, result.Header, "Accept-Encoding")

	//!TODO: fix requested range to be divisible by the storage partSize

	return result
}

func (h *reqHandler) getResponseHook() func(*httputils.FlexibleResponseWriter) {

	return func(rw *httputils.FlexibleResponseWriter) {
		h.Logger.Debugf("[%s] Received headers for %s, sending them to client...",
			h.reqID, h.req.URL)
		httputils.CopyHeadersWithout(rw.Headers, h.resp.Header(), hopHeaders...)
		h.resp.WriteHeader(rw.Code)

		isCacheable := cacheutils.IsResponseCacheable(rw.Code, rw.Headers)
		if !isCacheable {
			h.Logger.Debugf("[%s] Response is non-cacheable", h.reqID)
			rw.BodyWriter = utils.AddCloser(h.resp)
			return
		}

		expiresIn := cacheutils.ResponseExpiresIn(rw.Headers, h.CacheDefaultDuration)
		if expiresIn <= 0 {
			h.Logger.Debugf("[%s] Response expires in the past: %s", h.reqID, expiresIn)
			rw.BodyWriter = utils.AddCloser(h.resp)
			return
		}

		responseRange, err := httputils.GetResponseRange(rw.Code, rw.Headers)
		if err != nil {
			h.Logger.Debugf("[%s] Was not able to get response range (%s)",
				h.reqID, err)
			rw.BodyWriter = utils.AddCloser(h.resp)
			return
		}

		h.Logger.Debugf("[%s] Response is cacheable! Caching metadata and parts", h.reqID)

		code := rw.Code
		if code == http.StatusPartialContent {
			// 206 is returned only if the server would
			// have returned 200 with a normal request
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
		// maybe the server does not return date, we should set it then
		if obj.Headers.Get("Date") == "" {
			obj.Headers.Set("Date", now.Format(http.TimeFormat))
		}

		//!TODO: consult the cache algorithm whether to save the metadata
		//!TODO: optimize this, save the metadata only when it's newer
		//!TODO: also, error if we already have fresh metadata but the
		//       received metadata is different
		if err := h.Cache.Storage.SaveMetadata(obj); err != nil {
			h.Logger.Errorf("[%s] Could not save metadata for %s: %s",
				h.reqID, obj.ID, err)
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

		h.Logger.Debugf("[%s] Setting the cached data to expire in %s", h.reqID, expiresIn)
		h.Cache.Scheduler.AddEvent(
			h.objID.Hash(),
			storage.GetExpirationHandler(h.Cache, h.Logger, h.objID),
			expiresIn,
		)
	}
}

func idSuffix(s, e uint64) []byte {
	return strconv.AppendUint(append(strconv.AppendUint([]byte(`->b=`), s, 10), '-'), e, 10)
}

func (h *reqHandler) getUpstreamReader(start, end uint64) io.ReadCloser {
	subh := *h
	// ->start-end
	subh.ctx, subh.reqID = contexts.AppendToRequestID(subh.ctx, idSuffix(start, end))
	subh.req = subh.getNormalizedRequest()
	subh.req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", start, end))

	h.Logger.Debugf("[%s] Making upstream request for %s, bytes [%d-%d]...",
		subh.reqID, subh.req.URL, start, end)

	//!TODO: optimize requests for the same pieces?
	//       if possible, make only 1 request to the upstream for the same part

	r, w := io.Pipe()
	subh.resp = httputils.NewFlexibleResponseWriter(func(rw *httputils.FlexibleResponseWriter) {
		respRng, err := httputils.GetResponseRange(rw.Code, rw.Headers)
		if err != nil {
			h.Logger.Debugf("[%s] Could not parse the content-range"+
				"for the partial upstream request: %s",
				subh.reqID, err)
			_ = w.CloseWithError(err)
		}
		h.Logger.Debugf("[%s] Received response with status %d and range %v",
			subh.reqID, rw.Code, respRng)
		if rw.Code == http.StatusPartialContent {
			//!TODO: check whether the returned range corresponds to the requested range
			rw.BodyWriter = w
		} else if rw.Code == http.StatusOK {
			//!TODO: handle this, use skipWriter or something like that
			_ = w.CloseWithError(fmt.Errorf("NOT IMPLEMENTED"))
		} else {
			_ = w.CloseWithError(
				fmt.Errorf("Upstream responded with status %d", rw.Code))
		}
	})
	go utils.SafeExecute(
		subh.carbonCopyProxy,
		func(err error) {
			h.Logger.Errorf("[%s] Panic inside carbonCopyProxy %s", subh.reqID, err)
			w.CloseWithError(err) // !TODO maybe some other error
		},
	)
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
		h.Logger.Errorf("[%s] Unexpected error while trying to load %s from storage: %s",
			h.reqID, idx, err)
	} else if cached {
		h.Logger.Debugf(
			"[%s] CacheAlgorithm said a part %s is cached but Storage couldn't find it",
			h.reqID, idx)
	}
	return nil, nil
}

func (h *reqHandler) getContents(indexes []*types.ObjectIndex, from int,
) (io.ReadCloser, int, error) {
	r, err := h.getPartFromStorage(indexes[from])
	if r != nil {
		return r, 1, nil
	} else if err != nil {
		return nil, 0, err
	}

	partSize := h.Cache.Storage.PartSize()
	fromByte := uint64(indexes[from].Part) * partSize
	parts, err := h.Cache.Storage.GetAvailableParts(h.objID)
	if err != nil {
		return nil, 0, err
	}
	sort.Sort(objectIndexes(parts))
	i := sort.Search(len(parts), func(i int) bool {
		return (parts[i].Part > indexes[from].Part &&
			parts[i].Part <= indexes[len(indexes)-1].Part)
	})
	if i < len(parts) { // there is a part
		r, _ = h.getPartFromStorage(parts[i])
		if r != nil {
			toByte := umin(h.obj.Size, uint64(parts[i].Part)*partSize-1)
			return utils.MultiReadCloser(
					h.getUpstreamReader(fromByte, toByte), r),
				int(parts[i].Part-indexes[from].Part) + 1, nil
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
			h.Logger.Errorf(
				"[%s] Unexpected error while trying to load %s from storage: %s",
				h.reqID, indexes[i], err)
			return
		}
		if i == 0 && startOffset > 0 {
			contents, err = utils.SkipReadCloser(contents, int64(startOffset))
			if err != nil {
				h.Logger.Errorf(
					"[%s] Unexpected error while trying to skip %d from %s: %s",
					h.reqID, startOffset, indexes[i], err)
				return
			}
		}
		if i+partsCount == len(indexes) {
			endLimit := uint64(partsCount-1)*partSize + end%partSize + 1
			if i == 0 {
				endLimit -= startOffset
			}
			//!TODO: fix int conversion
			contents = utils.LimitReadCloser(contents, int64(endLimit))
		}

		if copied, err := io.Copy(h.resp, contents); err != nil {
			h.Logger.Logf(
				"[%s] Error sending contents after %dbytes of %s, parts[%d-%d]: %s",
				h.reqID, copied, h.objID, indexes[i].Part,
				indexes[i+partsCount-1].Part, err)

			shouldReturn = true
		}
		//!TODO: compare the copied length with the expected
		if err := contents.Close(); err != nil {
			h.Logger.Errorf(
				"[%s] Error while closing content reader for %s, parts[%d-%d]: %s",
				h.reqID, h.objID, indexes[i].Part, indexes[i+partsCount-1].Part,
				err)
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

type objectIndexes []*types.ObjectIndex

func (o objectIndexes) Len() int {
	return len(o)
}
func (o objectIndexes) Less(i, j int) bool {
	return o[i].Part < o[j].Part
}
func (o objectIndexes) Swap(i, j int) {
	o[i], o[j] = o[j], o[i]
}
