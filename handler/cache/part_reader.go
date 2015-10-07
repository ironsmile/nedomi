package cache

import (
	"io"
	"io/ioutil"

	"github.com/ironsmile/nedomi/utils"
)

// read until boundaries of parts
type partReadCloser struct {
	io.ReadCloser
	partSize  uint64
	readSoFar int
}

func newWholeChunkReadCloser(r io.ReadCloser, size uint64) io.ReadCloser {
	return &partReadCloser{
		ReadCloser: r,
		partSize:   size,
	}
}

func (f *partReadCloser) Read(b []byte) (int, error) {
	n, err := f.ReadCloser.Read(b)
	f.readSoFar += n
	return n, err
}

func (f *partReadCloser) Close() error {
	var readFromCurrent = uint64(f.readSoFar) % f.partSize
	var err error
	if readFromCurrent != 0 {
		_, err = io.CopyN(
			ioutil.Discard,
			f.ReadCloser,
			int64(f.partSize-readFromCurrent),
		)
	}
	if err == io.EOF {
		err = nil
	}

	return utils.NewCompositeError(err, f.ReadCloser.Close())
}
