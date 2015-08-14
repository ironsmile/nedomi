package simple

import (
	"fmt"
	"net/http"
	"net/url"
)

// Upstream is a basic HTTP upstream implementation.
type Upstream struct {
	client http.Client
	url    *url.URL
}

// New returns a configured and ready to use Upstream instance.
func New(url *url.URL) *Upstream {
	return &Upstream{
		client: http.Client{},
		url:    url,
	}
}

// GetRequest executes a simple GET HTTP request to the upstream server.
func (u *Upstream) GetRequest(pathStr string) (*http.Response, error) {
	newURL, err := u.createNewURL(pathStr)
	if err != nil {
		return nil, err
	}

	return u.client.Get(newURL.String())
}

// GetRequestPartial executes a GET HTTP request to the upstream server with a
// range header, specified by stand and end.
func (u *Upstream) GetRequestPartial(pathStr string, start, end uint64) (*http.Response, error) {
	newURL, err := u.createNewURL(pathStr)
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
func (u *Upstream) Head(pathStr string) (*http.Response, error) {
	newURL, err := u.createNewURL(pathStr)
	if err != nil {
		return nil, err
	}
	return u.client.Head(newURL.String())
}

// GetSize retrieves the file size of the specified path from the upstream server.
func (u *Upstream) GetSize(pathStr string) (int64, error) {
	resp, err := u.Head(pathStr)
	if err != nil {
		return 0, err
	}

	return resp.ContentLength, nil
}

// GetHeader retrieves the headers for the specified path from the upstream server.
func (u *Upstream) GetHeader(pathStr string) (*http.Response, error) {
	resp, err := u.Head(pathStr)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (u *Upstream) createNewURL(pathStr string) (*url.URL, error) {
	path, err := url.Parse(pathStr)
	if err != nil {
		return nil, err
	}

	return u.url.ResolveReference(path), nil
}
