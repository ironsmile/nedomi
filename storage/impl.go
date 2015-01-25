package storage

import (
	"io"
	"net/http"
	"os"
	"path"
	"strconv"
	"syscall"

	"github.com/gophergala/nedomi/cache"
	. "github.com/gophergala/nedomi/config"
	. "github.com/gophergala/nedomi/types"
)

type storageImpl struct {
	cache          cache.CacheManager
	partSize       uint64 // actually uint32
	storageObjects uint64
	path           string
}

func NewStorage(config CacheZoneSection) Storage {
	return &storageImpl{
		partSize:       config.PartSize.Bytes(),
		storageObjects: config.StorageObjects,
		path:           config.Path,
	}
}

func urlForId(id ObjectID) string {
	return (string)(id) // @todo redo
}

func pathFromID(id ObjectID) string {
	return (string)(id)
}

func pathFromIndex(index ObjectIndex) string {
	return path.Join(pathFromID(index.ObjID), strconv.Itoa(int(index.Part)))
}

func (s *storageImpl) Get(id ObjectID, start, end uint64) (io.ReadCloser, error) {
	indexes := s.breakInIndexes(id, start, end)
	readers := make([]io.ReadCloser, len(indexes))
	for i, index := range indexes {
		if s.cache.Has(index) {
			file, err := os.Open(s.pathFromIndex(index))
			if err != nil {
				return nil, err
			}
			readers[i] = file
		} else {
			readers[i] = s.newResponseReaderFor(index)
		}
		s.cache.UsedObjectIndex(index)
	}

	// work in start and end
	var startOffset, endLimit = start % s.partSize, end % s.partSize
	readers[0] = newSkipReadCloser(readers[0], int(startOffset))
	readers[len(readers)-1] = newLimitReadCloser(readers[len(readers)-1], int(endLimit))

	return newMultiReadCloser(readers...), nil
}

func (s *storageImpl) newResponseReaderFor(index ObjectIndex) io.ReadCloser {
	responseReader := &ResponseReader{
		done: make(chan struct{}),
	}
	go func() {
		resp, err := s.GetRequest(index)
		if err != nil {
			responseReader.SetErr(err)
		} else {
			defer resp.Body.Close()
			file_path := s.pathFromIndex(index)
			file, err := os.OpenFile(file_path, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0600)
			if err != nil {
				responseReader.SetErr(err)
			}
			_, err = io.Copy(file, resp.Body)
			if err != nil {
				responseReader.SetErr(err)
				file.Close()
			}
			if !s.cache.ObjectIndexStored(index) {
				err := syscall.Unlink(file_path)
				if err != nil {
					responseReader.SetErr(err)
					file.Close()
				}

			}

			_, err = file.Seek(0, os.SEEK_SET)
			if err != nil {
				responseReader.SetErr(err)
				file.Close()
			}
			responseReader.SetReadFrom(file)
		}
	}()
	return responseReader
}

type ResponseReader struct {
	readCloser io.ReadCloser
	done       chan struct{}
	err        error
}

func (r *ResponseReader) SetErr(err error) {
	r.err = err
	close(r.done)
}

func (r *ResponseReader) SetReadFrom(reader io.ReadCloser) {
	r.readCloser = reader
	close(r.done)
}

func (r *ResponseReader) Close() error {
	<-r.done
	if r.err != nil {
		return r.err
	}
	return r.readCloser.Close()
}
func (r *ResponseReader) Read(p []byte) (int, error) {
	<-r.done
	if r.err != nil {
		return 0, r.err
	}
	return r.readCloser.Read(p)
}

func (s *storageImpl) breakInIndexes(id ObjectID, start, end uint64) []ObjectIndex {
	firstIndex := start / s.partSize
	lastIndex := end / s.partSize
	result := make([]ObjectIndex, 0, lastIndex-firstIndex)
	for i := firstIndex; i < lastIndex; i++ {
		result = append(result, ObjectIndex{id, uint32(i)})
	}
	return result
}

func (s *storageImpl) pathFromIndex(index ObjectIndex) string {
	return path.Join(s.path, pathFromIndex(index))
}

func (s *storageImpl) pathFromID(id ObjectID) string {
	return path.Join(s.path, pathFromID(id))
}

func (s *storageImpl) Discard(id ObjectID) error {
	return os.RemoveAll(s.pathFromID(id))
}

func (s *storageImpl) DiscardIndex(index ObjectIndex) error {
	return os.Remove(path.Join(s.pathFromIndex(index)))
}

func (s *storageImpl) GetRequest(index ObjectIndex) (*http.Response, error) {
	return nil, nil
}
