package storage

import (
	"io"
	"net/http"
	"os"
	"time"

	"github.com/ironsmile/nedomi/config"
	"github.com/ironsmile/nedomi/types"
	"github.com/ironsmile/nedomi/utils"
	"golang.org/x/net/context"
)

//!TODO: move this as a simple handler - there is no need to "orchestrate" anything after all...

// Orchestrator is responsible for coordinating and synchronizing the operations
// between the Storage, CacheAlgorithm and each vhost's Upstream server.
type Orchestrator struct {
	cfg                 *config.CacheZone
	storage             types.Storage
	algorithm           types.CacheAlgorithm
	logger              types.Logger
	objectPartsToRemove chan *types.ObjectIndex
	done                <-chan struct{}
}

func (o *Orchestrator) getPartWriter(objID *types.ObjectID, startPos uint64) io.WriteCloser {
	return &partWriter{
		objID:    objID,
		partSize: o.cfg.PartSize.Bytes(),
		storage:  o.storage,
	}
}

func (o *Orchestrator) getPartReader(req *http.Request, objID *types.ObjectID) io.ReadCloser {
	//!TODO: handle ranges correctly, handle *unknown* object size correctly
	readers := []io.ReadCloser{}
	var part uint32
	for {
		idx := &types.ObjectIndex{ObjID: objID, Part: part}
		o.logger.Logf("[%p] Trying to load part %s from storage...", req, idx)
		r, err := o.storage.GetPart(idx)
		if err != nil {
			break //TODO: fix, this is wrong
		}
		o.logger.Logf("[%p] Loaded part %s from storage!", req, idx)
		readers = append(readers, r)
		part++
	}
	o.logger.Logf("[%p] Return reader with %d parts of %s from storage!", req, len(readers), objID)
	return MultiReadCloser(readers...)
}

func (o *Orchestrator) getResponseHook(resp http.ResponseWriter, req *http.Request,
	objID *types.ObjectID, saveMetadata bool) func(*utils.FlexibleResponseWriter) {

	return func(rw *utils.FlexibleResponseWriter) {
		o.logger.Logf("[%p] Received headers for %s, sending them to client...", req, req.URL)
		for h := range rw.Headers {
			resp.Header()[h] = rw.Headers[h]
		}
		resp.WriteHeader(rw.Code)

		//!TODO: handle duration
		isCacheable, _ := utils.IsResponseCacheable(rw.Code, rw.Headers)
		if !isCacheable {
			o.logger.Logf("[%p] Response is non-cacheable :(", req)
			rw.BodyWriter = NopCloser(resp)
			return
		}

		o.logger.Logf("[%p] Response is cacheable! Caching parts...", req)
		if saveMetadata {
			obj := &types.ObjectMetadata{
				ID:                objID,
				ResponseTimestamp: time.Now().Unix(),
				Code:              rw.Code,
				Headers:           rw.Headers,
			}
			if err := o.storage.SaveMetadata(obj); err != nil {
				o.logger.Errorf("Could not save metadata for %s: %s", obj.ID, err)
			}
		}

		//!TODO: handle missing content length, partial requests (range headers), etc.
		rw.BodyWriter = MultiWriteCloser(NopCloser(resp), o.getPartWriter(objID, 0))
	}
}

// Handle is used for serving requests for which the client could receive cached responses.
func (o *Orchestrator) Handle(ctx context.Context, resp http.ResponseWriter,
	req *http.Request, loc *types.Location) {

	objID := types.NewObjectID(loc.CacheKey, req.URL.String())
	o.logger.Logf("[%p] Handling request for %s by orchestrator...", req, req.URL)
	//spew.Dump(req.Header)

	obj, err := o.storage.GetMetadata(objID)
	if err == nil && isMetadataFresh(obj) {
		o.logger.Logf("[%p] Metadata found, downloading only missing parts...", req)

		for h := range obj.Headers {
			resp.Header()[h] = obj.Headers[h]
		}
		resp.WriteHeader(obj.Code)

		//!TODO: handle ranges correctly
		reader := o.getPartReader(req, objID)
		if copied, err := io.Copy(resp, reader); err != nil {
			o.logger.Errorf("[%p] Error copying response: %s. Copied %d", req, err, copied)
		}
		reader.Close()
		return
	}

	if err != nil && !os.IsNotExist(err) {
		o.logger.Logf("[%p] Storage error when reading metadata: %s", req, err)
	} else if !isMetadataFresh(obj) {
		o.logger.Logf("[%p] Metadata is stale, refreshing...", req)
		//!TODO: optimize, do only a head request when the metadata is stale
	}

	//!TODO: consult the cache algorithm whether to save the metadata
	hook := o.getResponseHook(resp, req, objID, true)
	flexibleResp := utils.NewFlexibleResponseWriter(hook)

	//!TODO: copy the request, do not use the original; modify other headers?
	req.Header.Del("Accept-Encoding")
	loc.Upstream.ServeHTTP(flexibleResp, req)
	flexibleResp.BodyWriter.Close()
}
