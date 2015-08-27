package storage

import (
	"fmt"
	"io"
	"net/http"

	"github.com/ironsmile/nedomi/cache"
	"github.com/ironsmile/nedomi/config"
	"github.com/ironsmile/nedomi/types"
	"golang.org/x/net/context"
)

// Orchestrator is responsible for coordinating and synchronizing the operations
// between the Storage, CacheAlgorithm and each vhost's Upstream server.
type Orchestrator struct {
	storage    types.Storage
	algorithm  types.CacheAlgorithm
	logger     types.Logger
	requestsCh chan *request
	removeCh   chan types.ObjectIndex
	doneCh     <-chan struct{}
}

type request struct {
	ctx context.Context
	req *http.Request
}

// Handle returns the object's metadata and an io.ReadCloser that will read from
// the `start` of the requested object to the `end`.
func (o *Orchestrator) Handle(ctx context.Context, req *http.Request) (*types.ObjectMetadata, io.ReadCloser) {

	//!TODO: use the loop and queue duplicate requests
	/*
		get into the loop with an initial request
		in the loop:
			if the storage does not have the metadata, proxy the request to the upstream, return the result to the client and record it in the cache
			if the storage has the metadata and it's fresh, return it and a lazy readcloser, while gathering the needed pieces from the storage and the upstream
	*/

	return nil, nil
}

// GetCacheStats returns the cache statistics.
func (o *Orchestrator) GetCacheStats() types.CacheStats {
	return o.algorithm.Stats()
}

func (o *Orchestrator) loop() {
	closing := false
	defer func() {
		// This is safe to do because algorithm.AddObject() is called only in
		// the loop and will not be called anymore.
		close(o.removeCh)
	}()

	for {
		select {
		case req := <-o.requestsCh:
			if req == nil {
				panic("request is nil")
			}

			if closing {
				//!TODO: implement proper closing sequence
				continue
			}
			/*
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

				case resp := <-upstreamResponses:
					// handle
			*/
		case oi := <-o.removeCh:
			//!TODO: if this is the last index, remove the whole object
			o.storage.DiscardPart(oi)

		case <-o.doneCh:
			o.logger.Log("Shutting down the storage orchestrator...")
			closing = true

		}
	}

}

// NewOrchestrator creates and initializes a new Orchestrator object and starts
// its scheduling goroutines.
func NewOrchestrator(ctx context.Context, cfg *config.CacheZoneSection,
	logger types.Logger) (o *Orchestrator, err error) {

	o = &Orchestrator{
		logger:   logger,
		removeCh: make(chan types.ObjectIndex, 1000),
		doneCh:   ctx.Done(),
	}

	// Initialize the cache algorithm
	if o.algorithm, err = cache.New(cfg, o.removeCh, logger); err != nil {
		return nil, fmt.Errorf("Could not initialize storage algorithm '%s': %s", cfg.Algorithm, err)
	}

	// Initialize the storage
	if o.storage, err = New(cfg, logger); err != nil {
		return nil, fmt.Errorf("Could not initialize storage '%s': %s", cfg.Type, err)
	}

	go o.loop()

	return o, nil
}
