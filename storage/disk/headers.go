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

func (s *Disk) readHeaderFromFile(id types.ObjectID) (http.Header, error) {
	file, err := os.Open(path.Join(s.path, headerFileNameFromID(id)))
	if err == nil {
		defer file.Close()
		var header http.Header
		err := json.NewDecoder(file).Decode(&header)
		return header, err

	} else if !os.IsNotExist(err) {
		s.logger.Errorf("Got error while trying to open headers file: %s", err)
	}
	return nil, err
}

func (s *Disk) writeHeaderToFile(id types.ObjectID, header http.Header) {
	file, err := CreateFile(path.Join(s.path, headerFileNameFromID(id)))
	if err != nil {
		s.logger.Errorf("Couldn't create file to write header: %s", err)
		return
	}

	defer file.Close()
	if err := json.NewEncoder(file).Encode(header); err != nil {
		s.logger.Errorf("Error while writing header to file: %s", err)
	}
}
*/
