package simple

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"
)

// New returns a configured and ready to use Upstream instance.
func New(upstream *url.URL) *httputil.ReverseProxy {
	director := func(req *http.Request) {
		req.URL.Scheme = upstream.Scheme
		req.URL.Host = upstream.Host
		req.Host = upstream.Host
	}

	//!TODO: record statistics (times, errors, etc.) for all requests
	return &httputil.ReverseProxy{
		FlushInterval: 200 * time.Millisecond, //!TODO: get from config
		Director:      director,
		//!TODO: set ErrorLog to our own logger
	}
}
