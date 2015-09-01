package disk

/*
func downloadHeaders(ctx context.Context, hq *headerQueue, finished chan<- *headerQueue) {
	vhost, ok := contexts.GetVhost(ctx)
	if !ok {
		hq.err = fmt.Errorf("Could not get vhost from context.")
		finished <- hq
		return
	}

	resp, err := vhost.Upstream.GetHeader(hq.id.Path)
	if err != nil {
		hq.err = err
	} else {
		hq.header = resp.Header
		//!TODO: handle allowed cache duration
		hq.isCacheable, _ = utils.IsResponseCacheable(resp)
	}

	finished <- hq
}

type headerRequest struct {
	id      types.ObjectID
	header  http.Header
	err     error
	done    chan struct{}
	context context.Context
}

type headerQueue struct {
	id          types.ObjectID
	header      http.Header
	isCacheable bool
	err         error
	requests    []*headerRequest
}

func newHeaderQueue(request *headerRequest) *headerQueue {
	return &headerQueue{
		id:       request.id,
		requests: []*headerRequest{request},
	}
}

*/
