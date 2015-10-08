package proxy

import (
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"time"

	"golang.org/x/net/context"

	"github.com/ironsmile/nedomi/config"
	"github.com/ironsmile/nedomi/types"
)

// Settings contains the possible settings for the proxy
type Settings struct {
}

// Proxy is a simple reverse proxy
type Proxy struct {
	*httputil.ReverseProxy
}

// New returns a configured and ready to use Upstream instance.
func New(cfg *config.Handler, l *types.Location, next types.RequestHandler) (*Proxy, error) {
	//!TODO: get this from the config
	if l.UpstreamAddress == nil {
		return nil, fmt.Errorf("No upstream address for proxy handler in %s", l.Name)
	}

	//!TODO: get all of these hardcoded values from the config
	director := func(req *http.Request) {
		req.URL.Scheme = l.UpstreamAddress.Scheme
		req.URL.Host = l.UpstreamAddress.Host
		req.Host = l.UpstreamAddress.Host

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
	return &Proxy{&httputil.ReverseProxy{
		FlushInterval: 200 * time.Millisecond,
		Director:      director,
		Transport:     transport,
		//!TODO: set ErrorLog to our own logger
	}}, nil
}

// RequestHandle implements the interface
func (p *Proxy) RequestHandle(_ context.Context, w http.ResponseWriter, r *http.Request) {
	p.ServeHTTP(w, r)
}
