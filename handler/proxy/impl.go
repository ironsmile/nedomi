package proxy

import (
	"io"
	"net"
	"net/http"
	"strings"

	"github.com/ironsmile/nedomi/types"
	"github.com/ironsmile/nedomi/utils/httputils"

	"golang.org/x/net/context"
)

// This code is based on the source of ReverseProxy from Go's standard library:
// https://golang.org/src/net/http/httputil/reverseproxy.go
// Copyright 2011 The Go Authors.

// ReverseProxy is an HTTP Handler that takes an incoming request and
// sends it to another server, proxying the response back to the
// client.
type ReverseProxy struct {
	// The transport used to perform proxy requests.
	// If nil, http.DefaultTransport is used.
	Transport http.RoundTripper

	// Logger specifies a logger for errors that occur when attempting
	// to proxy the request.
	Logger types.Logger
}

// RequestHandle implements the nedomi interface
func (p *ReverseProxy) RequestHandle(_ context.Context, w http.ResponseWriter, r *http.Request) {
	p.ServeHTTP(w, r)
}

// Hop-by-hop headers. These are removed when sent to the backend.
var hopHeaders = httputils.GetHopByHopHeaders()

type requestCanceler interface {
	CancelRequest(*http.Request)
}

type runOnFirstRead struct {
	io.Reader // optional; nil means empty body

	fn func() // Run before first Read, then set to nil
}

func (c *runOnFirstRead) Read(bs []byte) (int, error) {
	if c.fn != nil {
		c.fn()
		c.fn = nil
	}
	if c.Reader == nil {
		return 0, io.EOF
	}
	return c.Reader.Read(bs)
}

func (p *ReverseProxy) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	transport := p.Transport
	if transport == nil {
		transport = http.DefaultTransport
	}

	outreq := new(http.Request)
	*outreq = *req // includes shallow copies of maps, but okay
	httputils.CopyHeadersWithout(req.Header, outreq.Header, hopHeaders...)

	if closeNotifier, ok := rw.(http.CloseNotifier); ok {
		if requestCanceler, ok := transport.(requestCanceler); ok {
			reqDone := make(chan struct{})
			defer close(reqDone)

			clientGone := closeNotifier.CloseNotify()

			outreq.Body = struct {
				io.Reader
				io.Closer
			}{
				Reader: &runOnFirstRead{
					Reader: outreq.Body,
					fn: func() {
						go func() {
							select {
							case <-clientGone:
								requestCanceler.CancelRequest(outreq)
							case <-reqDone:
							}
						}()
					},
				},
				Closer: outreq.Body,
			}
		}
	}

	outreq.Proto = "HTTP/1.1"
	outreq.ProtoMajor = 1
	outreq.ProtoMinor = 1
	outreq.Close = false
	// If we don't set it, Go sets it for us to something stupid...
	outreq.Header.Set("User-Agent", "nedomi") //!TODO: get user-agent from config

	//!TODO: get host value from config
	//outreq.Host = l.UpstreamAddress.Host

	if clientIP, _, err := net.SplitHostPort(req.RemoteAddr); err == nil {
		// If we aren't the first proxy retain prior
		// X-Forwarded-For information as a comma+space
		// separated list and fold multiple headers into one.
		if prior, ok := outreq.Header["X-Forwarded-For"]; ok {
			clientIP = strings.Join(prior, ", ") + ", " + clientIP
		}
		outreq.Header.Set("X-Forwarded-For", clientIP)
	}

	res, err := transport.RoundTrip(outreq)
	if err != nil {
		p.Logger.Logf("[%p] Proxy error: %v", req, err)
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	for _, h := range hopHeaders {
		res.Header.Del(h)
	}

	httputils.CopyHeaders(res.Header, rw.Header())

	// The "Trailer" header isn't included in the Transport's response,
	// at least for *http.Transport. Build it up from Trailer.
	if len(res.Trailer) > 0 {
		var trailerKeys []string
		for k := range res.Trailer {
			trailerKeys = append(trailerKeys, k)
		}
		rw.Header().Add("Trailer", strings.Join(trailerKeys, ", "))
	}

	rw.WriteHeader(res.StatusCode)
	if len(res.Trailer) > 0 {
		// Force chunking if we saw a response trailer.
		// This prevents net/http from calculating the length for short
		// bodies and adding a Content-Length.
		if fl, ok := rw.(http.Flusher); ok {
			fl.Flush()
		}
	}
	io.Copy(rw, res.Body)
	res.Body.Close() // close now, instead of defer, to populate res.Trailer
	httputils.CopyHeaders(res.Trailer, rw.Header())
}
