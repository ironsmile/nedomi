package types

import (
	"net"
	"net/url"
)

// UpstreamAddress represents a resolved upstream address
type UpstreamAddress struct {
	URL    *url.URL
	IP     net.IP
	Port   int
	Weight float64
}
