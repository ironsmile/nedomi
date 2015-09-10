package storage

import (
	"io"

	"github.com/ironsmile/nedomi/utils"
)

type nopCloser struct {
	io.Writer
}

func (nopCloser) Close() error { return nil }

// NopCloser returns a WriteCloser with a no-op Close method wrapping
// the provided Writer w.
func NopCloser(w io.Writer) io.WriteCloser {
	return nopCloser{w}
}

type multiWriteCloser struct {
	writers []io.WriteCloser
}

func (t *multiWriteCloser) Write(p []byte) (n int, err error) {
	for _, w := range t.writers {
		n, err = w.Write(p)
		if err != nil {
			return
		}
		if n != len(p) {
			err = io.ErrShortWrite
			return
		}
	}
	return len(p), nil
}

func (t *multiWriteCloser) Close() error {
	errors := []error{}
	for _, w := range t.writers {
		errors = append(errors, w.Close())
	}
	return utils.NewCompositeError(errors...)
}

// MultiWriteCloser creates a writer that duplicates its writes to all the
// provided writers, similar to the Unix tee(1) command.
func MultiWriteCloser(writers ...io.WriteCloser) io.WriteCloser {
	w := make([]io.WriteCloser, len(writers))
	copy(w, writers)
	return &multiWriteCloser{w}
}