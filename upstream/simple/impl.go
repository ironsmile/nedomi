package simple

import (
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"
)

// New returns a configured and ready to use Upstream instance.
func New(upstream *url.URL) *httputil.ReverseProxy {
	//!TODO: after upstream is a simple handler, get all of these hardcoded
	// values from the config

	director := func(req *http.Request) {
		req.URL.Scheme = upstream.Scheme
		req.URL.Host = upstream.Host
		req.Host = upstream.Host

		// If we don't set it, Go sets it for us to something stupid...
		req.Header.Set("User-Agent", "nedomi")
	}

	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		Dial: (&net.Dialer{
			Timeout:   10 * time.Second,
			KeepAlive: 10 * time.Second,
		}).Dial,
		TLSHandshakeTimeout: 5 * time.Second,
		DisableKeepAlives:   false,
		DisableCompression:  true,
		MaxIdleConnsPerHost: 5,
	}

	//!TODO: record statistics (times, errors, etc.) for all requests
	return &httputil.ReverseProxy{
		FlushInterval: 200 * time.Millisecond,
		Director:      director,
		Transport:     transport,
		//!TODO: set ErrorLog to our own logger
	}
}
