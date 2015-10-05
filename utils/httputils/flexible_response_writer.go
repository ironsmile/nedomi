package httputils

import (
	"errors"
	"io"
	"net/http"
)

// FlexibleResponseWriter is an implementation of http.ResponseWriter that calls
// a hook function before accepting writes. The hook function's job is to
// inspect the current state and determine where the body should be written.
type FlexibleResponseWriter struct {
	Code        int         // the HTTP response code from WriteHeader
	Headers     http.Header // the HTTP response headers
	BodyWriter  io.WriteCloser
	hook        func(*FlexibleResponseWriter)
	wroteHeader bool
}

// NewFlexibleResponseWriter returns an initialized FlexibleResponseWriter.
func NewFlexibleResponseWriter(hook func(*FlexibleResponseWriter)) *FlexibleResponseWriter {
	return &FlexibleResponseWriter{
		Code:    200,
		Headers: make(http.Header),
		hook:    hook,
	}
}

// Header returns the response headers.
func (frw *FlexibleResponseWriter) Header() http.Header {
	if frw.Headers == nil {
		frw.Headers = make(http.Header)
	}
	return frw.Headers
}

// Write checks if a writer is initialized. If there is a body writer, it passes
// the arguments to it. If there isn't one, it fails.
func (frw *FlexibleResponseWriter) Write(buf []byte) (int, error) {
	if !frw.wroteHeader {
		frw.WriteHeader(frw.Code)
	}

	if frw.BodyWriter == nil {
		return 0, errors.New("The body is not initialized, writes are not accepted.")
	}
	return frw.BodyWriter.Write(buf)
}

// WriteHeader sets rw.Code and calls the hook function.
func (frw *FlexibleResponseWriter) WriteHeader(code int) {
	if frw.wroteHeader {
		return
	}
	frw.Code = code
	frw.wroteHeader = true
	frw.hook(frw)
}

// Close closes the internal bodyWriter
func (frw *FlexibleResponseWriter) Close() error {
	if frw.BodyWriter == nil {
		return nil
	}
	return frw.BodyWriter.Close()

}
