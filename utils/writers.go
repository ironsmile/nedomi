package utils

import "io"

type nopCloser struct {
	io.Writer
}

func (nopCloser) Close() error { return nil }

// NopCloser returns a WriteCloser with a no-op Close method wrapping
// the provided Writer w.
func NopCloser(w io.Writer) io.WriteCloser {
	return nopCloser{w}
}

// AddCloser adds io.Closer to a io.Writer.
// If the provided io.Writer is io.WriteCloser
// it's just casted, otherwise a NopCloser is used
func AddCloser(w io.Writer) io.WriteCloser {
	if wc, ok := w.(io.WriteCloser); ok {
		return wc
	}
	return NopCloser(w)
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
	return NewCompositeError(errors...)
}

// MultiWriteCloser creates a writer that duplicates its writes to all the
// provided writers, similar to the Unix tee(1) command.
func MultiWriteCloser(writers ...io.WriteCloser) io.WriteCloser {
	w := make([]io.WriteCloser, len(writers))
	copy(w, writers)
	return &multiWriteCloser{w}
}
