package types

import (
	"fmt"
	"net/url"
)

// UpstreamAddress represents a resolved upstream address
type UpstreamAddress struct {
	url.URL
	Hostname    string
	Port        string
	OriginalURL *url.URL
	Weight      uint32
}

func (ua *UpstreamAddress) String() string {
	return fmt.Sprintf("%s|%d (%s)", ua.URL.String(), ua.Weight, ua.OriginalURL)
}
