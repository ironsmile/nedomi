package storage

import (
	"io"
	"log"

	"github.com/ironsmile/nedomi/utils"
)

type multiReadCloser struct {
	readers []io.ReadCloser
	index   int
}

// MultiReadCloser returns a io.ReadCloser that's the logical concatenation of
// the provided input readers.
func MultiReadCloser(readerClosers ...io.ReadCloser) io.ReadCloser {

	return &multiReadCloser{
		readers: readerClosers,
	}
}

func (m *multiReadCloser) Read(p []byte) (int, error) {
	if m.index == len(m.readers) {
		return 0, io.EOF
	}

	size, err := m.readers[m.index].Read(p)
	if err != nil {
		if err != io.EOF {
			return size, err
		}
		if closeErr := m.readers[m.index].Close(); closeErr != nil {
			log.Printf("Got error while closing no longer needed readers inside multiReadCloser: %s\n", closeErr)
		}
		m.index++
		if m.index != len(m.readers) {
			err = nil
		}
	}

	return size, err
}

func (m *multiReadCloser) Close() error {
	c := new(utils.CompositeError)
	for ; m.index < len(m.readers); m.index++ {
		err := m.readers[m.index].Close()
		if err != nil {
			c.AppendError(err)
		}
	}

	if c.Empty() {
		c = nil
	}
	return c

}

type limitedReadCloser struct {
	io.ReadCloser
	maxLeft int
}

// LimitReadCloser wraps a io.ReadCloser but stops with EOF after `max` bytes.
func LimitReadCloser(readCloser io.ReadCloser, max int) io.ReadCloser {
	return &limitedReadCloser{
		ReadCloser: readCloser,
		maxLeft:    max,
	}
}

func (r *limitedReadCloser) Read(p []byte) (int, error) {
	readSize := min(r.maxLeft, len(p))
	size, err := r.ReadCloser.Read(p[:readSize])
	r.maxLeft -= size
	if r.maxLeft == 0 && err == nil {
		err = io.EOF
	}
	return size, err
}

func min(l, r int) int {
	if l > r {
		return r
	}
	return l
}

type skippingReadCloser struct {
	io.ReadCloser
	skipLeft int
}

// SkipReadCloser wraps a io.ReadCloser and ignores the first `skip` bytes.
func SkipReadCloser(readCloser io.ReadCloser, skip int) io.ReadCloser {
	return &skippingReadCloser{
		ReadCloser: readCloser,
		skipLeft:   skip,
	}
}

const skipBufSize = 512

var b [skipBufSize]byte

func (r *skippingReadCloser) Read(p []byte) (int, error) {
	for r.skipLeft > 0 {
		readSize := min(r.skipLeft, skipBufSize)
		size, err := r.ReadCloser.Read(b[:readSize])
		r.skipLeft -= size
		if err != nil {
			return 0, err
		}
	}

	return r.ReadCloser.Read(p)
}

// TeeReadCloser is a io.TeeReader with Close() support.
func TeeReadCloser(r io.ReadCloser, w io.Writer) io.ReadCloser {
	return &teeReadCloser{r, w}
}

type teeReadCloser struct {
	r io.ReadCloser
	w io.Writer
}

func (t *teeReadCloser) Read(p []byte) (n int, err error) {
	n, err = t.r.Read(p)
	if n > 0 {
		if n, err := t.w.Write(p[:n]); err != nil {
			return n, err
		}
	}
	return
}

func (t *teeReadCloser) Close() error {
	return t.r.Close()
}
