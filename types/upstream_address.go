package types

import "net/url"

// UpstreamAddress represents a resolved upstream address
type UpstreamAddress struct {
	URL         *url.URL
	ResolvedURL *url.URL
	Weight      float64
}
