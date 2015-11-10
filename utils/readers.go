package utils

import (
	"io"
	"io/ioutil"
	"log"
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

func (m *multiReadCloser) WriteTo(w io.Writer) (n int64, err error) {
	if m.index == len(m.readers) {
		return 0, io.EOF
	}

	var (
		nn  int64
		rrs = m.readers[m.index:]
	)

	for _, reader := range rrs {
		nn, err = io.Copy(w, reader)
		n += nn
		if err != nil {
			return
		}
		m.index++
	}
	return
}

func (m *multiReadCloser) Close() error {
	c := new(CompositeError)
	for ; m.index < len(m.readers); m.index++ {
		err := m.readers[m.index].Close()
		if err != nil {
			c.AppendError(err)
		}
	}

	if c.Empty() {
		return nil
	}
	return c

}

type limitedReadCloser struct {
	io.Reader
	io.Closer
}

// LimitReadCloser wraps a io.ReadCloser but stops with EOF after `max` bytes.
func LimitReadCloser(readCloser io.ReadCloser, max int64) io.ReadCloser {
	return &limitedReadCloser{
		Reader: io.LimitReader(readCloser, max),
		Closer: readCloser,
	}
}

func min(l, r int) int {
	if l > r {
		return r
	}
	return l
}

type skippingReadCloser struct {
	io.ReadCloser
	skip int64
}

func (lrc *limitedReadCloser) WriteTo(w io.Writer) (n int64, err error) {
	return io.Copy(w, lrc.Reader)
}

// SkipReadCloser wraps a io.ReadCloser and ignores the first `skip` bytes.
func SkipReadCloser(readCloser io.ReadCloser, skip int64) (result io.ReadCloser, err error) {
	if seeker, ok := readCloser.(io.Seeker); ok {
		_, err = seeker.Seek(skip, 1)
		if err == nil {
			result = readCloser
		}
		return
	}

	return &skippingReadCloser{
		ReadCloser: readCloser,
		skip:       skip,
	}, nil
}

func (r *skippingReadCloser) Read(p []byte) (int, error) {
	if r.skip > 0 {
		if n, err := io.CopyN(ioutil.Discard, r.ReadCloser, r.skip); err != nil {
			r.skip -= n
			return 0, err
		}
		r.skip = 0
	}

	return r.ReadCloser.Read(p)
}
