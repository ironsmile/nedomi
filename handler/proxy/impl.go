package proxy

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"

	"github.com/ironsmile/nedomi/contexts"
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
	// The transport used to perform upstream requests.
	defaultUpstream types.Upstream

	// Logger specifies a logger for errors that occur when attempting
	// to proxy the request.
	Logger types.Logger

	Settings Settings

	CodesToRetry map[int]string
}

// RequestHandle implements the nedomi interface
func (p *ReverseProxy) RequestHandle(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	p.ServeHTTP(ctx, w, r)
}

// Hop-by-hop headers. These are removed when sent to the backend.
var hopHeaders = httputils.GetHopByHopHeaders()

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

func (p *ReverseProxy) getOutRequest(reqID types.RequestID, rw http.ResponseWriter, req *http.Request, upstream types.Upstream) (*http.Request, error) {
	outreq := new(http.Request)
	*outreq = *req
	url := *req.URL
	outreq.URL = &url

	outreq.Header = http.Header{}
	httputils.CopyHeadersWithout(req.Header, outreq.Header, hopHeaders...)
	outreq.Header.Set("User-Agent", p.Settings.UserAgent) // If we don't set it, Go sets it for us to something stupid...

	outreq.RequestURI = ""
	outreq.Proto = "HTTP/1.1"
	outreq.ProtoMajor = 1
	outreq.ProtoMinor = 1
	outreq.Close = false

	upAddr, err := upstream.GetAddress(p.Settings.UpstreamHashPrefix + req.URL.Path)
	if err != nil {
		return nil, fmt.Errorf("[%s] Proxy handler could not get an upstream address: %v", reqID, err)
	}
	p.Logger.Debugf("[%s] Using upstream %s (%s) to proxy request", reqID, upAddr, upAddr.OriginalURL)
	outreq.URL.Scheme = upAddr.Scheme
	outreq.URL.Host = upAddr.Host
	outreq.URL.User = upAddr.User

	// Set the correct host
	if p.Settings.HostHeader != "" {
		outreq.Host = p.Settings.HostHeader
	} else if p.Settings.HostHeaderKeepOriginal {
		if req.Host != "" {
			outreq.Host = req.Host
		} else {
			outreq.Host = req.URL.Host
		}
	} else {
		outreq.Host = upAddr.OriginalURL.Host
	}

	if closeNotifier, ok := rw.(http.CloseNotifier); ok {
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
							upstream.CancelRequest(outreq)
						case <-reqDone:
						}
					}()
				},
			},
			Closer: outreq.Body,
		}
	}

	if clientIP, _, err := net.SplitHostPort(req.RemoteAddr); err == nil {
		// If we aren't the first proxy retain prior
		// X-Forwarded-For information as a comma+space
		// separated list and fold multiple headers into one.
		if prior, ok := outreq.Header["X-Forwarded-For"]; ok {
			clientIP = strings.Join(prior, ", ") + ", " + clientIP
		}
		outreq.Header.Set("X-Forwarded-For", clientIP)
	}

	return outreq, nil
}

func (p *ReverseProxy) doRequestFor(
	ctx context.Context,
	reqID types.RequestID,
	rw http.ResponseWriter,
	req *http.Request,
	upstream types.Upstream,
) (*http.Response, error) {
	outreq, err := p.getOutRequest(reqID, rw, req, upstream)
	if err != nil {
		return nil, err
	}

	return upstream.Do(ctx, outreq)
}

func (p *ReverseProxy) ServeHTTP(ctx context.Context, rw http.ResponseWriter, req *http.Request) {
	var upstream = p.defaultUpstream
	reqID, _ := contexts.GetRequestID(ctx)
	res, err := p.doRequestFor(ctx, reqID, rw, req, upstream)
	if err != nil {
		p.Logger.Logf("[%s] Proxy error: %v", reqID, err)
		httputils.Error(rw, http.StatusInternalServerError)
		return
	}
	if newUpstream, ok := p.CodesToRetry[res.StatusCode]; ok {
		upstream = getUpstreamFromContext(ctx, newUpstream)
		if upstream != nil {
			if err = res.Body.Close(); err != nil {
				p.Logger.Logf("[%s] Proxy error on closing response which will be retried: %v",
					reqID, err)
			}

			res, err = p.doRequestFor(ctx, reqID, rw, req, upstream)
			if err != nil {
				p.Logger.Logf("[%s] Proxy error: %v", reqID, err)
				httputils.Error(rw, http.StatusInternalServerError)
				return
			}
		} else {
			p.Logger.Errorf("[%s] Proxy was configured to retry on code %d with upstream %s but no such upstream exist",
				reqID, res.StatusCode, newUpstream)
		}
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
	if _, err := io.Copy(rw, res.Body); err != nil {
		p.Logger.Logf("[%s] Proxy error during copying: %v", reqID, err)
	}

	// Close now, instead of defer, to populate res.Trailer
	if err := res.Body.Close(); err != nil {
		p.Logger.Errorf("[%s] Proxy error during response close: %v", reqID, err)
	}
	httputils.CopyHeaders(res.Trailer, rw.Header())
}

func getUpstreamFromContext(ctx context.Context, upstream string) types.Upstream {
	app, ok := contexts.GetApp(ctx)
	if !ok {
		panic("no app in context") // seriosly ?
	}

	return app.GetUpstream(upstream)
}
