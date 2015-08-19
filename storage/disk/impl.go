package disk

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path"
	"strconv"
	"syscall"

	"github.com/ironsmile/nedomi/config"
	"github.com/ironsmile/nedomi/logger"
	"github.com/ironsmile/nedomi/types"
	"github.com/ironsmile/nedomi/utils"
)

const headerFileName = "headers"

// Disk implements the Storage interface by writing data to a disk
type Disk struct {
	cache          types.CacheAlgorithm
	partSize       uint64 // actually uint32
	storageObjects uint64
	path           string
	//!TODO: remove hardcoded single upstream, it can be different for each request
	upstream       types.Upstream
	indexRequests  chan *indexRequest
	headerRequests chan *headerRequest
	downloaded     chan *indexDownload
	removeChan     chan removeRequest
	logger         logger.Logger
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
func New(config config.CacheZoneSection, cm types.CacheAlgorithm,
	up types.Upstream, log logger.Logger) *Disk {
	storage := &Disk{
		partSize:       config.PartSize.Bytes(),
		storageObjects: config.StorageObjects,
		path:           config.Path,
		cache:          cm,
		upstream:       up,
		indexRequests:  make(chan *indexRequest),
		downloaded:     make(chan *indexDownload),
		removeChan:     make(chan removeRequest),
		headerRequests: make(chan *headerRequest),
		logger:         log,
	}

	err := os.RemoveAll(storage.path)
	if err != nil {
		storage.logger.Errorf("Couldn't clean path '%s', got error: %s", storage.path, err)
	}

	go storage.loop()

	return storage
}

func (s *Disk) downloadIndex(index types.ObjectIndex) (*os.File, *http.Response, error) {
	startOffset := uint64(index.Part) * s.partSize
	endOffset := startOffset + s.partSize - 1
	resp, err := s.upstream.GetRequestPartial(index.ObjID.Path, startOffset, endOffset)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()
	filePath := pathFromIndex(s.path, index)

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
	go func(download *indexDownload, index types.ObjectIndex) {
		file, resp, err := s.downloadIndex(index)
		if err != nil {
			download.err = err
		} else {
			download.file = file
			//!TODO: handle allowed cache duration
			download.isCacheable, _ = utils.IsResponseCacheable(resp)
		}
		s.downloaded <- download
	}(download, request.index)
	return download
}

func (s *Disk) loop() {
	downloading := make(map[types.ObjectIndex]*indexDownload)
	headers := make(map[types.ObjectID]*headerQueue)
	headerFinished := make(chan *headerQueue)
	for {
		select {
		case request := <-s.indexRequests:
			if request == nil {
				panic("request is nil")
			}
			s.logger.Debugf("Storage [%p]: downloading for indexRequest %+v\n", s, request)
			if download, ok := downloading[request.index]; ok {
				download.requests = append(download.requests, request)
				continue
			}
			if s.cache.Lookup(request.index) {
				file, err := os.Open(pathFromIndex(s.path, request.index))
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

		case request := <-s.removeChan:
			s.logger.Debugf("Storage [%p] removing %s", s, request.path)
			request.err <- syscall.Unlink(request.path)
			close(request.err)

		// HEADERS
		case request := <-s.headerRequests:
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

				go func(hq *headerQueue) {
					resp, err := s.upstream.GetHeader(hq.id.Path)
					if err != nil {
						hq.err = err
					} else {
						hq.header = resp.Header
						//!TODO: handle allowed cache duration
						hq.isCacheable, _ = utils.IsResponseCacheable(resp)
					}

					headerFinished <- hq
				}(queue)
			}

		case finished := <-headerFinished:
			delete(headers, finished.id)
			if finished.err == nil {
				if finished.isCacheable {
					//!TODO: do not save directly, run through the cache algo?
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
		}
	}
}
func (s *Disk) readHeaderFromFile(id types.ObjectID) (http.Header, error) {
	file, err := os.Open(path.Join(pathFromID(s.path, id), headerFileName))
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
	filePath := path.Join(pathFromID(s.path, id), headerFileName)
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
	id     types.ObjectID
	header http.Header
	err    error
	done   chan struct{}
}

type indexRequest struct {
	index  types.ObjectIndex
	reader io.ReadCloser
	err    error
	done   chan struct{}
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
func (s *Disk) GetFullFile(id types.ObjectID) (io.ReadCloser, error) {
	size, err := s.upstream.GetSize(id.Path)
	if err != nil {
		return nil, err
	}
	if size <= 0 {
		resp, err := s.upstream.GetRequest(id.Path)
		if err != nil {
			return nil, err
		}

		return resp.Body, nil
	}

	return s.Get(id, 0, uint64(size))
}

// Headers retunrs just the Headers for the specfied ObjectID
func (s *Disk) Headers(id types.ObjectID) (http.Header, error) {
	request := &headerRequest{
		id:   id,
		done: make(chan struct{}),
	}
	s.headerRequests <- request
	<-request.done
	return request.header, request.err
}

// Get retuns an ObjectID from start to end
func (s *Disk) Get(id types.ObjectID, start, end uint64) (io.ReadCloser, error) {
	indexes := breakInIndexes(id, start, end, s.partSize)
	readers := make([]io.ReadCloser, len(indexes))
	for i, index := range indexes {
		request := &indexRequest{
			index: index,
			done:  make(chan struct{}),
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

func pathFromIndex(root string, index types.ObjectIndex) string {
	return path.Join(pathFromID(root, index.ObjID), strconv.Itoa(int(index.Part)))
}

func pathFromID(root string, id types.ObjectID) string {
	return path.Join(root, id.CacheKey, id.Path)
}

type removeRequest struct {
	path string
	err  chan error
}

// Discard a previosly cached ObjectID
func (s *Disk) Discard(id types.ObjectID) error {
	request := removeRequest{
		path: pathFromID(s.path, id),
		err:  make(chan error),
	}

	s.removeChan <- request
	return <-request.err
}

// DiscardIndex a previosly cached ObjectIndex
func (s *Disk) DiscardIndex(index types.ObjectIndex) error {
	request := removeRequest{
		path: pathFromIndex(s.path, index),
		err:  make(chan error),
	}

	s.removeChan <- request
	return <-request.err
}

// GetCacheAlgorithm returns the used cache algorithm
func (s *Disk) GetCacheAlgorithm() *types.CacheAlgorithm {
	return &s.cache
}
