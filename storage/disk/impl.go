package disk

import (
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"strconv"
	"sync"
	"syscall"

	"github.com/ironsmile/nedomi/cache"
	"github.com/ironsmile/nedomi/config"
	"github.com/ironsmile/nedomi/types"
	"github.com/ironsmile/nedomi/upstream"
)

type storageImpl struct {
	cache          cache.Manager
	partSize       uint64 // actually uint32
	storageObjects uint64
	path           string
	upstream       upstream.Upstream
	indexRequests  chan *indexRequest
	downloaded     chan *indexDownload
	downloading    map[types.ObjectIndex]*indexDownload
	removeChan     chan removeRequest

	// The headers map must be guarded by a mutex. The storage object
	// is accessed in different goroutines and possibly threads. Without
	// the lock strange crashes may happen.
	metaHeadersLock sync.Mutex
	metaHeaders     map[types.ObjectID]http.Header
}

// New returns a new disk storage that ready for use.
func New(config config.CacheZoneSection, cm cache.Manager,
	up upstream.Upstream) *storageImpl {
	storage := &storageImpl{
		partSize:       config.PartSize.Bytes(),
		storageObjects: config.StorageObjects,
		path:           config.Path,
		cache:          cm,
		upstream:       up,
		metaHeaders:    make(map[types.ObjectID]http.Header),
		indexRequests:  make(chan *indexRequest),
		downloaded:     make(chan *indexDownload),
		downloading:    make(map[types.ObjectIndex]*indexDownload),
		removeChan:     make(chan removeRequest),
	}

	go storage.loop()

	return storage
}

func (s *storageImpl) downloadIndex(index types.ObjectIndex, vh *config.VirtualHost) (*os.File, error) {
	startOffset := uint64(index.Part) * s.partSize
	endOffset := startOffset + s.partSize - 1
	resp, err := s.upstream.GetRequestPartial(vh, index.ObjID.Path, startOffset, endOffset)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	filePath := s.pathFromIndex(index)

	if err := os.MkdirAll(path.Dir(filePath), 0700); err != nil {
		return nil, err
	}

	file, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE, 0600)
	if err != nil {
		file.Close()
		return nil, err
	}
	size, err := io.Copy(file, resp.Body)
	if err != nil {
		file.Close()
		return nil, err
	}
	log.Printf("Storage [%p] downloaded for index %s with size %d", s, index, size)

	_, err = file.Seek(0, os.SEEK_SET)
	if err != nil {
		file.Close()
		return nil, err
	}

	return file, err
}

func (s *storageImpl) startDownloadIndex(request *indexRequest) *indexDownload {
	download := &indexDownload{
		index:    request.index,
		requests: []*indexRequest{request},
	}
	go func(download *indexDownload, index types.ObjectIndex, vh *config.VirtualHost) {
		file, err := s.downloadIndex(index, vh)
		if err != nil {
			download.err = err
		} else {
			download.file = file
		}
		s.downloaded <- download
	}(download, request.index, request.vh)
	return download
}

type indexDownload struct {
	file     *os.File
	index    types.ObjectIndex
	err      error
	requests []*indexRequest
}

func (s *storageImpl) download(request *indexRequest) {
	log.Printf("Storage [%p]: downloading for indexRequest %+v\n", s, request)
	if download, ok := s.downloading[request.index]; ok {
		download.requests = append(download.requests, request)
	}
	s.startDownloadIndex(request)
}

func (s *storageImpl) loop() {
	for {
		select {
		case request := <-s.indexRequests:
			if s.cache.Lookup(request.index) {
				file, err := os.Open(s.pathFromIndex(request.index))
				if err != nil {
					log.Printf("Error while opening file in cache: %s", err)
					s.download(request)
				} else {
					request.reader = file
					s.cache.PromoteObject(request.index)
					close(request.done)
				}
			} else {
				s.download(request)
			}

		case download := <-s.downloaded:
			keep := s.cache.ShouldKeep(download.index)
			for _, request := range download.requests {
				if download.err != nil {
					log.Printf("Storage [%p]: error in downloading indexRequest %+v: %s\n", s, request, download.err)
					request.err = download.err
					close(request.done)
				} else {
					var err error
					request.reader, err = os.Open(download.file.Name()) // optimize
					s.cache.PromoteObject(request.index)
					if err != nil {
						log.Printf("Storage [%p]: error on reopening just downloaded file for indexRequest %+v :%s\n", s, request, err)
					}
					close(request.done)
				}
			}
			if !keep {
				syscall.Unlink(download.file.Name())
			}
		case request := <-s.removeChan:
			log.Printf("Storage [%p] removing %s", s, request.path)
			request.err <- syscall.Unlink(request.path)

		}
	}
}

type indexRequest struct {
	index  types.ObjectIndex
	vh     *config.VirtualHost
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

func (s *storageImpl) GetFullFile(vh *config.VirtualHost, id types.ObjectID) (io.ReadCloser, error) {
	size, err := s.upstream.GetSize(vh, id.Path)
	if err != nil {
		return nil, err
	}
	if size <= 0 {
		resp, err := s.upstream.GetRequest(vh, id.Path)
		if err != nil {
			return nil, err
		}

		return resp.Body, nil
	}

	return s.Get(vh, id, 0, uint64(size))
}

func (s *storageImpl) Headers(vh *config.VirtualHost, id types.ObjectID) (http.Header, error) {

	s.metaHeadersLock.Lock()
	defer s.metaHeadersLock.Unlock()

	if headers, ok := s.metaHeaders[id]; ok {
		return headers, nil
	}

	// Release the lock for the time of the HTTP request
	s.metaHeadersLock.Unlock()
	headers, err := s.upstream.GetHeader(vh, id.Path)
	// Get the lock back so that the deferred Unlock will not panic or something.
	// Also we need it in order to add an element to the headers map.
	s.metaHeadersLock.Lock()

	if err != nil {
		return nil, err
	}
	s.metaHeaders[id] = headers
	return headers, nil
}

func (s *storageImpl) Get(vh *config.VirtualHost, id types.ObjectID, start, end uint64) (io.ReadCloser, error) {
	indexes := s.breakInIndexes(id, start, end)
	readers := make([]io.ReadCloser, len(indexes))
	for i, index := range indexes {
		request := &indexRequest{
			index: index,
			done:  make(chan struct{}),
			vh:    vh,
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

func (s *storageImpl) breakInIndexes(id types.ObjectID, start, end uint64) []types.ObjectIndex {
	firstIndex := start / s.partSize
	lastIndex := end/s.partSize + 1
	result := make([]types.ObjectIndex, 0, lastIndex-firstIndex)
	for i := firstIndex; i < lastIndex; i++ {
		result = append(result, types.ObjectIndex{
			ObjID: id,
			Part:  uint32(i),
		})
	}
	return result
}

func (s *storageImpl) pathFromIndex(index types.ObjectIndex) string {
	return path.Join(s.pathFromID(index.ObjID), strconv.Itoa(int(index.Part)))
}

func (s *storageImpl) pathFromID(id types.ObjectID) string {
	return path.Join(s.path, id.CacheKey, id.Path)
}

type removeRequest struct {
	path string
	err  chan error
}

func (s *storageImpl) Discard(id types.ObjectID) error {
	request := removeRequest{
		path: s.pathFromID(id),
		err:  make(chan error),
	}

	s.removeChan <- request
	return <-request.err
}

func (s *storageImpl) DiscardIndex(index types.ObjectIndex) error {
	request := removeRequest{
		path: s.pathFromIndex(index),
		err:  make(chan error),
	}

	s.removeChan <- request
	return <-request.err
}
