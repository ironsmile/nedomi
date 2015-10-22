package upstream

import (
	"fmt"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/ironsmile/nedomi/config"
	"github.com/ironsmile/nedomi/types"
	"github.com/ironsmile/nedomi/upstream/balancing"
	"github.com/ironsmile/nedomi/utils/httputils"
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

	// Feed the unresolved addresses while waiting for DNS resolver
	unresolved := make([]*types.UpstreamAddress, len(conf.Addresses))
	for i, addr := range conf.Addresses {
		host, port, err := httputils.ParseURLHost(addr.URL)
		if err != nil {
			return nil, fmt.Errorf("Invalid upstream address %s: %s", addr.URL, err)
		}

		unresolved[i] = &types.UpstreamAddress{
			URL:         *addr.URL,
			Hostname:    host,
			Port:        port,
			OriginalURL: addr.URL,
			Weight:      addr.Weight,
		}
	}
	balancingAlgo.Set(unresolved)

	if conf.Settings.ResolveAddresses {
		//!TODO: get app cancel channel to the dns resolver
		go up.initDNSResolver(balancingAlgo, unresolved)
	}

	return up, nil
}

// NewSimple creates a simple RoundTripper with the default configuration that
// proxies requests to the supplied URL
func NewSimple(url *url.URL) (*Upstream, error) {
	host, port, err := httputils.ParseURLHost(url)
	if err != nil {
		return nil, fmt.Errorf("Invalid upstream address %s: %s", url, err)
	}

	up := &types.UpstreamAddress{
		URL:         *url,
		Hostname:    host,
		Port:        port,
		OriginalURL: url,
		Weight:      1,
	}

	return &Upstream{
		upTransport: getTransport(config.GetDefaultUpstreamSettings()),
		addressGetter: func(_ string) (*types.UpstreamAddress, error) {
			// Always return the same single url - no balancing needed
			return up, nil
		},
	}, nil
}
