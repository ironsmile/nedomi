package disk

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"strconv"
	"syscall"

	"golang.org/x/net/context"

	"github.com/ironsmile/nedomi/config"
	"github.com/ironsmile/nedomi/contexts"
	"github.com/ironsmile/nedomi/types"
	"github.com/ironsmile/nedomi/utils"
)

const (
	headerFileName   = "headers"
	objectIDFileName = "objID"
)

// Disk implements the Storage interface by writing data to a disk
type Disk struct {
	cache          types.CacheAlgorithm
	partSize       uint64 // actually uint32
	storageObjects uint64
	path           string
	indexRequests  chan *indexRequest
	headerRequests chan *headerRequest
	downloaded     chan *indexDownload
	removeChan     chan removeRequest
	logger         types.Logger
	closeCh        chan struct{}
}

type headerQueue struct {
	id          types.ObjectID
	header      http.Header
	isCacheable bool
	err         error
	requests    []*headerRequest
}

type indexDownload struct {
	file        *os.File
	isCacheable bool
	index       types.ObjectIndex
	err         error
	requests    []*indexRequest
}

// New returns a new disk storage that ready for use.
func New(config config.CacheZoneSection, ca types.CacheAlgorithm,
	log types.Logger) *Disk {
	storage := &Disk{
		partSize:       config.PartSize.Bytes(),
		storageObjects: config.StorageObjects,
		path:           config.Path,
		cache:          ca,
		indexRequests:  make(chan *indexRequest),
		downloaded:     make(chan *indexDownload),
		removeChan:     make(chan removeRequest),
		headerRequests: make(chan *headerRequest),
		closeCh:        make(chan struct{}),
		logger:         log,
	}

	go storage.loop()
	if err := storage.loadFromDisk(); err != nil {
		storage.logger.Error(err)
	}
	if err := storage.saveMetaToDisk(); err != nil {
		storage.logger.Error(err)
	}

	return storage
}

func (s *Disk) downloadIndex(ctx context.Context, index types.ObjectIndex) (*os.File, *http.Response, error) {

	vhost, ok := contexts.GetVhost(ctx)
	if !ok {
		return nil, nil, fmt.Errorf("Could not get vhost from context.")
	}

	startOffset := uint64(index.Part) * s.partSize
	endOffset := startOffset + s.partSize - 1
	resp, err := vhost.Upstream.GetRequestPartial(index.ObjID.Path, startOffset, endOffset)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()
	filePath := path.Join(s.path, pathFromIndex(index))

	if err := os.MkdirAll(path.Dir(filePath), 0700); err != nil {
		return nil, nil, err
	}

	file, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE, 0600)

	if err != nil {
		file.Close()
		return nil, nil, err
	}

	size, err := io.Copy(file, resp.Body)
	if err != nil {
		file.Close()
		return nil, nil, err
	}
	s.logger.Debugf("Storage [%p] downloaded for index %s with size %d", s, index, size)

	_, err = file.Seek(0, os.SEEK_SET)
	if err != nil {
		file.Close()
		return nil, nil, err
	}

	return file, resp, err
}

func (s *Disk) startDownloadIndex(request *indexRequest) *indexDownload {
	download := &indexDownload{
		index:    request.index,
		requests: []*indexRequest{request},
	}
	go func(ctx context.Context, download *indexDownload, index types.ObjectIndex) {
		file, resp, err := s.downloadIndex(ctx, index)
		if err != nil {
			download.err = err
		} else {
			download.file = file
			//!TODO: handle allowed cache duration
			download.isCacheable, _ = utils.IsResponseCacheable(resp)
			if download.isCacheable {
				s.writeObjectIdIfMissing(download.index.ObjID)
				//!TODO: don't do it for each piece and sanitize the headers
				s.writeHeaderToFile(download.index.ObjID, resp.Header)
			}
		}
		s.downloaded <- download
	}(request.context, download, request.index)
	return download
}

func (s *Disk) loop() {
	downloading := make(map[types.ObjectIndex]*indexDownload)
	headers := make(map[types.ObjectID]*headerQueue)
	headerFinished := make(chan *headerQueue)
	closing := false
	defer func() {
		close(headerFinished)
		close(s.downloaded)
		close(s.closeCh)
	}()
	for {
		select {
		case request := <-s.indexRequests:
			if closing {
				continue
			}

			if request == nil {
				panic("request is nil")
			}
			s.logger.Debugf("Storage [%p]: downloading for indexRequest %+v\n", s, request)
			if download, ok := downloading[request.index]; ok {
				download.requests = append(download.requests, request)
				continue
			}
			if s.cache.Lookup(request.index) {
				file, err := os.Open(path.Join(s.path, pathFromIndex(request.index)))
				if err != nil {
					s.logger.Errorf("Error while opening file in cache: %s", err)
					downloading[request.index] = s.startDownloadIndex(request)
				} else {
					request.reader = file
					s.cache.PromoteObject(request.index)
					close(request.done)
				}
			} else {
				downloading[request.index] = s.startDownloadIndex(request)
			}

		case download := <-s.downloaded:
			delete(downloading, download.index)

			for _, request := range download.requests {
				if download.err != nil {
					s.logger.Errorf("Storage [%p]: error in downloading indexRequest %+v: %s\n", s, request, download.err)
					request.err = download.err
					close(request.done)
				} else {
					var err error
					request.reader, err = os.Open(download.file.Name()) //!TODO: optimize
					if err != nil {
						s.logger.Errorf("Storage [%p]: error on reopening just downloaded file for indexRequest %+v :%s\n", s, request, err)
						request.err = err
					}
					if download.isCacheable {
						s.cache.PromoteObject(request.index)
					}
					close(request.done)
				}
			}
			if !download.isCacheable || !s.cache.ShouldKeep(download.index) {
				syscall.Unlink(download.file.Name())
			}
			if closing && len(headers) == 0 && len(downloading) == 0 {
				return
			}

		case request := <-s.removeChan:
			if closing {
				continue
			}

			s.logger.Debugf("Storage [%p] removing %s", s, request.path)
			request.err <- syscall.Unlink(request.path)
			close(request.err)

		// HEADERS
		case request := <-s.headerRequests:
			if closing {
				continue
			}

			if queue, ok := headers[request.id]; ok {
				queue.requests = append(queue.requests, request)
			} else {
				header, err := s.readHeaderFromFile(request.id)
				if err == nil {
					request.header = header
					request.err = err
					close(request.done)
					continue

				}

				queue := &headerQueue{
					id:       request.id,
					requests: []*headerRequest{request},
				}
				headers[request.id] = queue

				go func(ctx context.Context, hq *headerQueue) {
					vhost, ok := contexts.GetVhost(ctx)
					if !ok {
						hq.err = fmt.Errorf("Could not get vhost from context.")
						headerFinished <- hq
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

					headerFinished <- hq
				}(request.context, queue)
			}

		case finished := <-headerFinished:
			delete(headers, finished.id)
			if finished.err == nil {
				if finished.isCacheable {
					//!TODO: do not save directly, run through the cache algo?
					s.writeObjectIdIfMissing(finished.id)
					s.writeHeaderToFile(finished.id, finished.header)
				}
			}
			for _, request := range finished.requests {
				if finished.err != nil {
					request.err = finished.err
				} else {
					request.header = finished.header // @todo copy ?
				}
				close(request.done)
			}

			if closing && len(headers) == 0 && len(downloading) == 0 {
				return
			}

		case <-s.closeCh:
			closing = true
			if len(headers) == 0 && len(downloading) == 0 {
				return
			}
		}
	}
}

//Writes the ObjectID to the disk in it's place if it already hasn't been written
func (s *Disk) writeObjectIdIfMissing(id types.ObjectID) error {
	pathToObjectID := path.Join(s.path, objectIDFileNameFromID(id))
	if err := os.MkdirAll(path.Dir(pathToObjectID), 0700); err != nil {
		s.logger.Errorf("Couldn't make directory for ObjectID [%s]: %s", id, err)
		return err
	}

	file, err := os.OpenFile(pathToObjectID, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		if os.IsExist(err) {
			return nil
		}
		return err
	}
	defer file.Close()

	return json.NewEncoder(file).Encode(id)
}

func (s *Disk) readObjectIDForKeyNHash(key, hash string) (types.ObjectID, error) {
	s.logger.Errorf("Calling readObjectIDForKeyNHash(%s, %s)", key, hash)
	filePath := path.Join(s.path, key, hash, objectIDFileName)
	file, err := os.Open(filePath)
	if err != nil {
		return types.ObjectID{}, err
	}
	var id types.ObjectID
	if err := json.NewDecoder(file).Decode(&id); err != nil {
		return types.ObjectID{}, err
	}

	return id, nil
}

func objectIDFileNameFromID(id types.ObjectID) string {
	return path.Join(pathFromID(id), objectIDFileName)
}

func headerFileNameFromID(id types.ObjectID) string {
	return path.Join(pathFromID(id), headerFileName)
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
	filePath := path.Join(s.path, headerFileNameFromID(id))
	if err := os.MkdirAll(path.Dir(filePath), 0700); err != nil {
		s.logger.Errorf("Couldn't make directory for header file: %s", err)
		return
	}

	file, err := os.Create(filePath)
	defer file.Close()
	if err != nil {
		s.logger.Errorf("Couldn't create file to write header: %s", err)
		return
	}
	if err := json.NewEncoder(file).Encode(header); err != nil {
		s.logger.Errorf("Error while writing header to file: %s", err)
	}
}

type headerRequest struct {
	id      types.ObjectID
	header  http.Header
	err     error
	done    chan struct{}
	context context.Context
}

type indexRequest struct {
	index   types.ObjectIndex
	reader  io.ReadCloser
	err     error
	done    chan struct{}
	context context.Context
}

func (ir *indexRequest) Close() error {
	<-ir.done
	if ir.err != nil {
		return ir.err
	}
	return ir.reader.Close()
}
func (ir *indexRequest) Read(p []byte) (int, error) {
	<-ir.done
	if ir.err != nil {
		return 0, ir.err
	}
	return ir.reader.Read(p)
}

// GetFullFile returns the whole file specified by the ObjectID
func (s *Disk) GetFullFile(ctx context.Context, id types.ObjectID) (io.ReadCloser, error) {
	vhost, ok := contexts.GetVhost(ctx)
	if !ok {
		return nil, fmt.Errorf("Could not get vhost from context.")
	}

	size, err := vhost.Upstream.GetSize(id.Path)
	if err != nil {
		return nil, err
	}
	if size <= 0 {
		resp, err := vhost.Upstream.GetRequest(id.Path)
		if err != nil {
			return nil, err
		}

		return resp.Body, nil
	}

	return s.Get(ctx, id, 0, uint64(size))
}

// Headers retunrs just the Headers for the specfied ObjectID
func (s *Disk) Headers(ctx context.Context, id types.ObjectID) (http.Header, error) {
	request := &headerRequest{
		id:      id,
		done:    make(chan struct{}),
		context: ctx,
	}
	s.headerRequests <- request
	<-request.done
	return request.header, request.err
}

// Get retuns an ObjectID from start to end
func (s *Disk) Get(ctx context.Context, id types.ObjectID, start, end uint64) (io.ReadCloser, error) {
	indexes := breakInIndexes(id, start, end, s.partSize)
	readers := make([]io.ReadCloser, len(indexes))
	for i, index := range indexes {
		request := &indexRequest{
			index:   index,
			done:    make(chan struct{}),
			context: ctx,
		}
		s.indexRequests <- request
		readers[i] = request
	}

	// work in start and end
	var startOffset, endLimit = start % s.partSize, end%s.partSize + 1
	readers[0] = newSkipReadCloser(readers[0], int(startOffset))
	readers[len(readers)-1] = newLimitReadCloser(readers[len(readers)-1], int(endLimit))

	return newMultiReadCloser(readers...), nil
}

func breakInIndexes(id types.ObjectID, start, end, partSize uint64) []types.ObjectIndex {
	firstIndex := start / partSize
	lastIndex := end/partSize + 1
	result := make([]types.ObjectIndex, 0, lastIndex-firstIndex)
	for i := firstIndex; i < lastIndex; i++ {
		result = append(result, types.ObjectIndex{
			ObjID: id,
			Part:  uint32(i),
		})
	}
	return result
}

func pathFromIndex(index types.ObjectIndex) string {
	return path.Join(pathFromID(index.ObjID), strconv.Itoa(int(index.Part)))
}

func pathFromID(id types.ObjectID) string {
	h := md5.New()
	io.WriteString(h, id.Path)
	return path.Join(id.CacheKey, hex.EncodeToString(h.Sum(nil)))
}

type removeRequest struct {
	path string
	err  chan error
}

// Discard a previosly cached ObjectID
func (s *Disk) Discard(id types.ObjectID) error {
	request := removeRequest{
		path: path.Join(s.path, pathFromID(id)),
		err:  make(chan error),
	}

	s.removeChan <- request
	return <-request.err
}

// DiscardIndex a previosly cached ObjectIndex
func (s *Disk) DiscardIndex(index types.ObjectIndex) error {
	request := removeRequest{
		path: path.Join(s.path, pathFromIndex(index)),
		err:  make(chan error),
	}

	s.removeChan <- request
	return <-request.err
}

// GetCacheAlgorithm returns the used cache algorithm
func (s *Disk) GetCacheAlgorithm() *types.CacheAlgorithm {
	return &s.cache
}

// Closes the Storage
func (s *Disk) Close() error {
	s.closeCh <- struct{}{}
	<-s.closeCh
	close(s.indexRequests)
	close(s.headerRequests)
	return nil
}
