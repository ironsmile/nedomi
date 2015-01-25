package storage

import "io"

type multiReadCloser struct {
	readers []io.ReadCloser
	index   int
}

func newMultiReadCloser(readerClosers ...io.ReadCloser) io.ReadCloser {

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
		m.index += 1
		if m.index != len(m.readers) {
			err = nil
		}
	}

	return size, err
}

func (m *multiReadCloser) Close() error {
	for _, reader := range m.readers {
		err := reader.Close()
		if err != nil {
			return err
		}
	}
	return nil

}

type limitedReadCloser struct {
	io.ReadCloser
	maxLeft int
}

func newLimitReadCloser(readCloser io.ReadCloser, max int) io.ReadCloser {
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
	} else {
		return l
	}
}

type skippingReadCloser struct {
	io.ReadCloser
	skipLeft int
}

func newSkipReadCloser(readCloser io.ReadCloser, skip int) io.ReadCloser {
	return &skippingReadCloser{
		ReadCloser: readCloser,
		skipLeft:   skip,
	}
}

func (r *skippingReadCloser) Read(p []byte) (int, error) {
	for r.skipLeft > 0 {
		var b [512]byte
		size, err := io.ReadAtLeast(r, b[:], r.skipLeft)
		if err != nil {
			return 0, err
		}
		r.skipLeft -= size

	}
	return r.ReadCloser.Read(p)
}
