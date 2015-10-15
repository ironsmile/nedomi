package upstream

import (
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/ironsmile/nedomi/config"
	"github.com/ironsmile/nedomi/types"
	"github.com/ironsmile/nedomi/upstream/balancing"
)

type upTransport interface {
	http.RoundTripper
	CancelRequest(*http.Request)
}

// Upstream implements the http.RoundTripper interface and is used for requests
// to all simple and advanced upstreams.
type Upstream struct {
	upTransport
	config        *config.Upstream
	logger        types.Logger
	addressGetter func(string) (*types.UpstreamAddress, error)
}

// GetAddress implements the Upstream interface
func (u *Upstream) GetAddress(uri string) (*types.UpstreamAddress, error) {
	return u.addressGetter(uri)
}

func getTransport(settings config.UpstreamSettings) upTransport {
	//!TODO: get all of these hardcoded values from the config
	//!TODO: use the facebook retryable transport
	//!TODO: investigate transport timeouts for active connections
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

	if settings.MaxConnectionsPerServer > 0 {
		return newConnectionLimiter(transport, settings.MaxConnectionsPerServer)
	}
	return transport
}

// New creates a new RoundTripper from the supplied upstream config
func New(conf *config.Upstream, logger types.Logger) (*Upstream, error) {

	balancingAlgo, err := balancing.New(conf.Balancing)
	if err != nil {
		return nil, err
	}

	up := &Upstream{
		upTransport:   getTransport(conf.Settings),
		config:        conf,
		logger:        logger,
		addressGetter: balancingAlgo.Get,
	}

	//!TODO: get app cancel channel to the dns resolver
	go up.initDNSResolver(balancingAlgo)

	return up, nil
}

// NewSimple creates a simple RoundTripper with the default configuration that
// proxies requests to the supplied URL
func NewSimple(url *url.URL) *Upstream {
	return &Upstream{
		upTransport: getTransport(config.GetDefaultUpstreamSettings()),
		addressGetter: func(_ string) (*types.UpstreamAddress, error) {
			// Always return the same single url - no balancing needed
			return &types.UpstreamAddress{URL: url, ResolvedURL: url, Weight: 1.0}, nil
		},
	}
}
