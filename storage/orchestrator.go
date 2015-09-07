package storage

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/ironsmile/nedomi/cache"
	"github.com/ironsmile/nedomi/config"
	"github.com/ironsmile/nedomi/contexts"
	"github.com/ironsmile/nedomi/types"
	"golang.org/x/net/context"
)

// Orchestrator is responsible for coordinating and synchronizing the operations
// between the Storage, CacheAlgorithm and each vhost's Upstream server.
type Orchestrator struct {
	cfg                 *config.CacheZoneSection
	storage             types.Storage
	algorithm           types.CacheAlgorithm
	logger              types.Logger
	clientRequests      chan *request
	objectPartsToRemove chan *types.ObjectIndex
	foundObjects        chan *storageItem
	done                <-chan struct{}
}

type request struct {
	ctx    context.Context
	req    *http.Request
	done   chan struct{}
	err    error
	obj    *types.ObjectMetadata
	reader io.ReadCloser
}

func (r *request) Close() error {
	<-r.done
	if r.err != nil {
		return r.err
	}
	return r.reader.Close()
}

func (r *request) Read(p []byte) (int, error) {
	<-r.done
	if r.err != nil {
		return 0, r.err
	}
	return r.reader.Read(p)
}

// Handle returns the object's metadata and an io.ReadCloser that will read from
// the `start` of the requested object to the `end`.
func (o *Orchestrator) Handle(ctx context.Context, req *http.Request) (*types.ObjectMetadata, io.ReadCloser, error) {

	orchReq := &request{
		req:  req,
		ctx:  ctx,
		done: make(chan struct{}),
	}
	select {
	case o.clientRequests <- orchReq:
	case <-o.done:
		return nil, nil, fmt.Errorf("Orchestrator shutdown")
	}

	<-orchReq.done
	return orchReq.obj, orchReq, orchReq.err

	//!TODO: use the loop and queue duplicate requests
	/*
		get into the loop with an initial request
		in the loop:
			if the storage does not have the metadata, proxy the request to the upstream, return the result to the client and record it in the cache
			if the storage has the metadata and it's fresh, return it and a lazy readcloser, while gathering the needed pieces from the storage and the upstream
	*/
}

func isMetadataFresh(obj *types.ObjectMetadata) bool {
	//!TODO: implement
	return true
}

func (o *Orchestrator) proxyRequestToUpstream(r *request) {

}

func (o *Orchestrator) handleClientRequest(r *request) {
	if r == nil {
		panic("request is nil")
	}
	l, ok := contexts.GetLocation(r.ctx)
	if !ok {
		r.err = fmt.Errorf("Could not get location from context")
		close(r.done)
		return
	}

	o.logger.Logf("[%p] Handling request for %s by orchestrator...", r.req, r.req.URL)

	resp := httptest.NewRecorder()
	l.Upstream.ServeHTTP(resp, r.req)

	r.obj = &types.ObjectMetadata{
		ID:                types.NewObjectID(l.CacheKey, r.req.URL.String()),
		ResponseTimestamp: time.Now().Unix(),
		Size:              uint64(resp.Body.Len()),
		Code:              resp.Code,
		Headers:           resp.HeaderMap,
	}
	r.reader = ioutil.NopCloser(resp.Body)
	close(r.done)
	/*

		obj, err := o.storage.GetMetadata(id)
		if os.IsNotExist(err) {
			o.logger.Logf("[%p] Metadata is not in storage, downloading...", r.req)
			//!TODO
		} else if err != nil {
			o.logger.Logf("[%p] Storage error when retrieving metadata: %s", r.req, err)
			//!TODO
		} else if !isMetadataFresh(obj) {
			o.logger.Logf("[%p] Metadata is stale, refreshing...", r.req)
			//!TODO
		}
	*/
	//!TODO: returns the metadata and create a lazy reader that tries to get all pieces

	// check storage for metadata
	// if present:
	//		object if is not cacheable
	//			directly proxy the request to the upstream, do not cache the result
	//		else if fresh:
	//			return the metadata and a lazily loaded ReadCloser
	//			for every piece that is present in the storage, open the file and queue it in the reader
	//			for every nonpresent chunk of pieces, send a request to the upstream and queue the created readers
	//		esle if not fresh, but otherwise suffictient to handle the request (future optimization):
	//			send a HEAD to the upstream
	//		else:
	//			fallthrough (*)
	// if not: (*)
	//		proxy the whole request to the Upstream
	//		if there are multiple request that are the same, they wait
	//		if there are diffent ranges, send them
}

func (o *Orchestrator) handleFoundObject(item *storageItem) {

}

func (o *Orchestrator) handleObjectPartRemoval(idx *types.ObjectIndex) {
	//!TODO: if this is the last index, remove the whole object
	//o.storage.DiscardPart(oi)

}

func (o *Orchestrator) loop() {
	//!TODO: for better performance, we can launch multiple scheduling loop(),
	// each taking of only a single cacheKey

	defer func() {
		// This is safe to do because algorithm.AddObject() is called only in
		// the loop and will not be called anymore.
		close(o.objectPartsToRemove)
	}()

	for {
		select {
		case req := <-o.clientRequests:
			o.handleClientRequest(req)

		case item := <-o.foundObjects:
			o.handleFoundObject(item)

		case idx := <-o.objectPartsToRemove:
			o.handleObjectPartRemoval(idx)

		case <-o.done:
			//!TODO: implement proper closing sequence
			o.logger.Log("Shutting down the storage orchestrator...")
			return
		}
	}

}

// GetCacheStats returns the cache statistics.
func (o *Orchestrator) GetCacheStats() types.CacheStats {
	return o.algorithm.Stats()
}

// NewOrchestrator creates and initializes a new Orchestrator object and starts
// its scheduling goroutines.
func NewOrchestrator(ctx context.Context, cfg *config.CacheZoneSection,
	logger types.Logger) (o *Orchestrator, err error) {

	o = &Orchestrator{
		cfg:                 cfg,
		logger:              logger,
		clientRequests:      make(chan *request),
		objectPartsToRemove: make(chan *types.ObjectIndex, 1000),
		foundObjects:        make(chan *storageItem),
		done:                ctx.Done(),
	}

	// Initialize the cache algorithm
	if o.algorithm, err = cache.New(cfg, o.objectPartsToRemove, logger); err != nil {
		return nil, fmt.Errorf("Could not initialize storage algorithm '%s': %s", cfg.Algorithm, err)
	}

	// Initialize the storage
	if o.storage, err = New(cfg, logger); err != nil {
		return nil, fmt.Errorf("Could not initialize storage '%s': %s", cfg.ID, err)
	}

	go o.loop()
	o.startConcurrentIterator()

	return o, nil
}
