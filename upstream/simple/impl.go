package simple

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/ironsmile/nedomi/config"
)

// SimpleUpstream is a basic HTTP upstream implementation. It recongizes how to
// make upstream requests by using the virtual host argument.
type SimpleUpstream struct {
	client http.Client
	cfg    *config.Config
}

// New returns a configured and ready to use SimpleUpstream instance.
func New(cfg *config.Config) *SimpleUpstream {
	return &SimpleUpstream{
		client: http.Client{},
		cfg:    cfg,
	}
}

// GetRequest executes a simple GET HTTP request to the upstream server.
func (u *SimpleUpstream) GetRequest(vh *config.VirtualHost, pathStr string) (*http.Response, error) {
	newURL, err := u.createNewURL(vh, pathStr)
	if err != nil {
		return nil, err
	}

	return u.client.Get(newURL.String())
}

// GetRequestPartial executes a GET HTTP request to the upstream server with a
// range header, specified by stand and end.
func (u *SimpleUpstream) GetRequestPartial(vh *config.VirtualHost,
	pathStr string, start, end uint64) (*http.Response, error) {
	newURL, err := u.createNewURL(vh, pathStr)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("GET", newURL.String(), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Range", fmt.Sprintf("bytes=%d-%d", start, end))
	return u.client.Do(req)
}

// Head executes a HEAD HTTP request to the upstream server.
func (u *SimpleUpstream) Head(vh *config.VirtualHost, pathStr string) (*http.Response, error) {
	newURL, err := u.createNewURL(vh, pathStr)
	if err != nil {
		return nil, err
	}
	return u.client.Head(newURL.String())
}

// GetSize retrieves the file size of the specified path from the upstream server.
func (u *SimpleUpstream) GetSize(vh *config.VirtualHost, pathStr string) (int64, error) {
	resp, err := u.Head(vh, pathStr)
	if err != nil {
		return 0, err
	}

	return resp.ContentLength, nil
}

// GetHeader retrieves the headers for the specified path from the upstream server.
func (u *SimpleUpstream) GetHeader(vh *config.VirtualHost, pathStr string) (http.Header, error) {
	resp, err := u.Head(vh, pathStr)
	if err != nil {
		return nil, err
	}

	return resp.Header, nil
}

func (u *SimpleUpstream) createNewURL(vh *config.VirtualHost, pathStr string) (*url.URL, error) {
	path, err := url.Parse(pathStr)
	if err != nil {
		return nil, err
	}

	return vh.UpstreamURL().ResolveReference(path), nil
}
