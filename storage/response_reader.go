package storage

import "io"

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
