package utils

import (
	"bytes"
	"net/http"
)

// DuplicatingResponseWriter is an implementation of http.ResponseWriter that
// passes all operations and data to another http.ResponseWriter instance and
// saves a copy of the data at the same time. It exports a *streaming*
// io.ReadCloser that can be used to read the response body while it's being
// flushed.
type DuplicatingResponseWriter struct {
	http.ResponseWriter
	Code int
	Body *bytes.Buffer
}

//!TODO: use useful stuff from golang.org/src/net/http/httptest/recorder.go like flush

// Write records the data and passes it to the original Write method.
func (drw *DuplicatingResponseWriter) Write(data []byte) (int, error) {
	drw.Body.Write(data)
	return drw.ResponseWriter.Write(data)
}

// WriteHeader saves the status code and calls the original WriteHeader method.
func (drw *DuplicatingResponseWriter) WriteHeader(code int) {
	drw.Code = code
	drw.ResponseWriter.WriteHeader(code)
}

// NewDuplicatingResponseWriter creates a new DuplicatingResponseWriter
func NewDuplicatingResponseWriter(sub http.ResponseWriter) *DuplicatingResponseWriter {
	return &DuplicatingResponseWriter{
		ResponseWriter: sub,
		Body:           bytes.NewBuffer([]byte{}),
	}
}
