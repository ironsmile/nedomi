package storage

import "io"

type ResponseReader struct {
	readCloser io.ReadCloser
	done       chan struct{}
	err        error
	before     func(*ResponseReader)
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
	if r.readCloser == nil && r.err == nil {
		r.before(r)
	}
	<-r.done
	if r.err != nil {
		return r.err
	}
	return r.readCloser.Close()
}
func (r *ResponseReader) Read(p []byte) (int, error) {
	if r.readCloser == nil && r.err == nil {
		r.before(r)
	}
	<-r.done
	if r.err != nil {
		return 0, r.err
	}
	return r.readCloser.Read(p)
}
