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
	"github.com/gophergala/nedomi/upstream"
)

type storageImpl struct {
	cache          cache.CacheManager
	partSize       uint64 // actually uint32
	storageObjects uint64
	path           string
	upstream       upstream.Upstream
}

func NewStorage(config CacheZoneSection, cm cache.CacheManager) Storage {
	return &storageImpl{
		partSize:       config.PartSize.Bytes(),
		storageObjects: config.StorageObjects,
		path:           config.Path,
		cache:          cm,
	}
}
func (s *storageImpl) GetFullFile(id ObjectID) (io.ReadCloser, error) {
	return nil, nil
}

func (s *storageImpl) Headers(id ObjectID) (http.Header, error) {
	return make(http.Header), nil
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
		startOffset := uint64(index.Part) * s.partSize
		endOffset := startOffset + s.partSize
		resp, err := s.upstream.GetRequest(s.pathFromIndex(index), startOffset, endOffset)
		if err != nil {
			responseReader.SetErr(err)
			return
		}
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
	}()
	return responseReader
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
	return path.Join(s.path, s.pathFromID(index.ObjID), strconv.Itoa(int(index.Part)))
}

func (s *storageImpl) pathFromID(id ObjectID) string {
	return path.Join(s.path, id.CacheKey, id.Path)
}

func (s *storageImpl) Discard(id ObjectID) error {
	return os.RemoveAll(s.pathFromID(id))
}

func (s *storageImpl) DiscardIndex(index ObjectIndex) error {
	return os.Remove(path.Join(s.pathFromIndex(index)))
}
