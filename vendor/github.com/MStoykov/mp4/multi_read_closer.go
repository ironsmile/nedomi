package mp4

import (
	"fmt"
	"io"
)

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
	if 0 == len(m.readers) {
		return 0, io.EOF
	}

	size, err := m.readers[0].Read(p)
	if err != nil {
		if err != io.EOF {
			return size, err
		}
		if closeErr := m.readers[0].Close(); closeErr != nil {
			fmt.Printf("Got error while closing no longer needed readers inside multiReadCloser: %s\n", closeErr)
		}
		m.readers = m.readers[1:]
		if 0 != len(m.readers) {
			err = nil
		}
	}

	return size, err
}

func (m *multiReadCloser) Close() error {
	var err error
	for _, reader := range m.readers {
		err = AppendError(err, reader.Close())
	}
	m.readers = nil

	return err
}

// AppendError appends one error to the other and doing the right thing if one or more of them are nil
func AppendError(err1 error, err2 error) error {
	if err2 == nil {
		return err1
	}
	if err1 == nil {
		return err2
	}
	return fmt.Errorf("%s after %s", err2, err1)
}
