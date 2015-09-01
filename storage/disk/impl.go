package disk

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"

	"github.com/ironsmile/nedomi/config"
	"github.com/ironsmile/nedomi/types"
	"github.com/ironsmile/nedomi/utils"
)

// Disk implements the Storage interface by writing data to a disk
type Disk struct {
	partSize        uint64
	path            string
	dirPermissions  os.FileMode
	filePermissions os.FileMode
	logger          types.Logger
}

// GetMetadata returns the metadata on disk for this object, if present.
func (s *Disk) GetMetadata(id *types.ObjectID) (*types.ObjectMetadata, error) {
	//!TODO: optimize - reading and parsing the file from disk every time is very ineffictient
	s.logger.Debugf("[DiskStorage] Getting metadata for %s...", id)
	return s.getObjectMetadata(s.getObjectMetadataPath(id))
}

// GetPart returns an io.ReadCloser that will read the specified part of the
// object from the disk.
func (s *Disk) GetPart(idx *types.ObjectIndex) (io.ReadCloser, error) {
	s.logger.Debugf("[DiskStorage] Getting file data for %s...", idx)
	f, err := os.Open(s.getObjectIndexPath(idx))
	if err != nil {
		return nil, err
	}
	stat, err := f.Stat()
	if err != nil {
		return nil, utils.NewCompositeError(err, f.Close())
	}

	if uint64(stat.Size()) > s.partSize {
		err = fmt.Errorf("Object part has invalid size %d", stat.Size())
		return nil, utils.NewCompositeError(err, f.Close(), s.DiscardPart(idx))
	}

	return f, nil
}

// SaveMetadata writes the supplied metadata to the disk.
func (s *Disk) SaveMetadata(m *types.ObjectMetadata) error {
	s.logger.Debugf("[DiskStorage] Saving metadata for %s...", m.ID)
	f, err := s.createFile(s.getObjectMetadataPath(m.ID))
	if err != nil {
		return err
	}

	//!TODO: use a faster encoding than json (some binary marshaller? gob?)
	if err = json.NewEncoder(f).Encode(m); err != nil {
		return utils.NewCompositeError(err, f.Close())
	}
	return f.Close()
}

// SavePart writes the contents of the supplied object part to the disk.
func (s *Disk) SavePart(idx *types.ObjectIndex, data io.Reader) error {
	s.logger.Debugf("[DiskStorage] Saving file data for %s...", idx)

	if _, err := os.Stat(s.getObjectMetadataPath(idx.ObjID)); err != nil {
		return fmt.Errorf("Could not read metadata file: %s", err)
	}

	f, err := s.createFile(s.getObjectIndexPath(idx))
	if err != nil {
		return err
	}

	savedSize, err := io.Copy(f, data)
	if err != nil {
		return utils.NewCompositeError(err, f.Close(), s.DiscardPart(idx))
	}

	if uint64(savedSize) > s.partSize {
		err = fmt.Errorf("Object part has invalid size %d", savedSize)
		return utils.NewCompositeError(err, f.Close(), s.DiscardPart(idx))
	}

	return f.Close()
}

// Discard removes the object and its metadata from the disk.
func (s *Disk) Discard(id *types.ObjectID) error {
	s.logger.Debugf("[DiskStorage] Discarding %s...", id)
	return os.RemoveAll(s.getObjectIDPath(id))
}

// DiscardPart removes the specified part of an Object from the disk.
func (s *Disk) DiscardPart(idx *types.ObjectIndex) error {
	s.logger.Debugf("[DiskStorage] Discarding %s...", idx)
	return os.Remove(s.getObjectIndexPath(idx))
}

// Iterate is a disk-specific function that iterates over all the objects on the
// disk and passes them to the supplied callback function. If the callback
// function returns false, the iteration stops.
func (s *Disk) Iterate(callback func(*types.ObjectMetadata, types.ObjectIndexMap) bool) error {
	// At most count(cacheKeys)*256*256 directories
	rootDirs, err := filepath.Glob(s.path + "/*/[0-9a-f][0-9a-f]/[0-9a-f][0-9a-f]")
	if err != nil {
		return err
	}

	for _, rootDir := range rootDirs {
		objectDirs, err := ioutil.ReadDir(rootDir)
		if err != nil {
			return err
		}

		for _, objectDir := range objectDirs {
			objectDirPath := path.Join(rootDir, objectDir.Name(), objectMetadataFileName)
			obj, err := s.getObjectMetadata(objectDirPath)
			if err != nil {
				return err
			}
			parts, err := s.getAvailableParts(obj)
			if err != nil {
				return err
			}
			if !callback(obj, parts) {
				return nil
			}
		}
	}
	return nil
}

// New returns a new disk storage that ready for use.
func New(cfg *config.CacheZoneSection, log types.Logger) (*Disk, error) {
	if cfg == nil || log == nil {
		return nil, fmt.Errorf("Nil constructor parameters")
	}

	if cfg.PartSize == 0 {
		return nil, fmt.Errorf("Invalid partSize value")
	}

	if _, err := os.Stat(cfg.Path); err != nil {
		return nil, fmt.Errorf("Cannot stat the disk storage path %s: %s", cfg.Path, err)
	}

	s := &Disk{
		partSize:        cfg.PartSize.Bytes(),
		path:            cfg.Path,
		dirPermissions:  0700 | os.ModeDir, //!TODO: get from the config
		filePermissions: 0600,              //!TODO: get from the config
		logger:          log,
	}

	testFilePath := path.Join(s.path, "constructorTestFile")
	if f, err := os.OpenFile(testFilePath,
		os.O_CREATE|os.O_EXCL|os.O_WRONLY, s.filePermissions); err != nil {
		return nil, fmt.Errorf("Could not write in the specified path %s: %s", s.path, err)
	} else if err := f.Close(); err != nil {
		return nil, fmt.Errorf("Could not close test file %s: %s", testFilePath, err)
	} else if err := os.Remove(testFilePath); err != nil {
		return nil, fmt.Errorf("Could not remove test file %s: %s", testFilePath, err)
	}

	return s, nil
}

/*
type indexDownload struct {
	file        *os.File
	isCacheable bool
	index       types.ObjectIndex
	err         error
	requests    []*indexRequest
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
	file, err := CreateFile(path.Join(s.path, pathFromIndex(index)))
	if err != nil {
		return nil, nil, err
	}

	size, err := io.Copy(file, resp.Body)
	if err != nil {
		return nil, nil, utils.NewCompositeError(err, file.Close())
	}
	s.logger.Debugf("Storage [%p] downloaded for index %s with size %d", s, index, size)

	_, err = file.Seek(0, os.SEEK_SET)
	if err != nil {
		return nil, nil, utils.NewCompositeError(err, file.Close())
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
				s.writeObjectIDIfMissing(download.index.ObjID)
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
			s.logger.Debugf("Storage [%p] removing %s", s, request.path)
			request.err <- syscall.Unlink(request.path)
			close(request.err)

		// HEADERS
		case request := <-s.headerRequests:
			if queue, ok := headers[request.id]; ok {
				queue.requests = append(queue.requests, request)
				continue
			}
			header, err := s.readHeaderFromFile(request.id)
			if err == nil {
				request.header = header
				close(request.done)
				continue
			}
			//!TODO handle error

			queue := newHeaderQueue(request)
			headers[request.id] = queue
			go downloadHeaders(request.context, queue, headerFinished)

		case finished := <-headerFinished:
			delete(headers, finished.id)
			if finished.err == nil {
				if finished.isCacheable {
					//!TODO: do not save directly, run through the cache algo?
					s.writeObjectIDIfMissing(finished.id)
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
			close(s.indexRequests)
			s.indexRequests = nil
			close(s.headerRequests)
			s.headerRequests = nil
			close(s.removeChan)
			s.removeChan = nil
			if len(headers) == 0 && len(downloading) == 0 {
				return
			}
		}
	}
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

// OldGet retuns an ObjectID from start to end
func (s *Disk) OldGet(ctx context.Context, id types.ObjectID, start, end uint64) (io.ReadCloser, error) {
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
	lastIndex := end/partSize + 1	//!TODO: FIX for sizes that are exact multiples of partSize
	result := make([]types.ObjectIndex, 0, lastIndex-firstIndex)
	for i := firstIndex; i < lastIndex; i++ {
		result = append(result, types.ObjectIndex{
			ObjID: id,
			Part:  uint32(i),
		})
	}
	return result
}

type removeRequest struct {
	path string
	err  chan error
}

// OldDiscard a previosly cached ObjectID
func (s *Disk) OldDiscard(id types.ObjectID) error {
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

// Close shuts down the Storage
func (s *Disk) Close() error {
	s.closeCh <- struct{}{}
	<-s.closeCh
	return nil
}
*/
