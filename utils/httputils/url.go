package httputils

import (
	"fmt"
	"net"
	"net/url"
	"strings"
)

// ParseURLHost extracts the hostname (without port) and the port from the
// supplied URL
func ParseURLHost(u *url.URL) (host, port string, err error) {
	if strings.ContainsRune(u.Host, ':') && !strings.HasSuffix(u.Host, "]") {
		return net.SplitHostPort(u.Host)
	}

	scheme := strings.ToLower(u.Scheme)
	if scheme == "http" {
		return net.SplitHostPort(u.Host + ":80")
	} else if scheme == "https" {
		return net.SplitHostPort(u.Host + ":443")
	}

	return u.Host, "", fmt.Errorf("address %s has an invalid scheme", u)
}
