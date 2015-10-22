package types

import "net/url"

// UpstreamAddress represents a resolved upstream address
type UpstreamAddress struct {
	url.URL
	Hostname    string
	Port        string
	OriginalURL *url.URL
	Weight      uint32
}
