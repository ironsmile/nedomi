package simple

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/ironsmile/nedomi/config"
)

type SimpleUpstream struct {
	client http.Client
	cfg    *config.Config
}

func New(cfg *config.Config) *SimpleUpstream {
	return &SimpleUpstream{
		client: http.Client{},
		cfg:    cfg,
	}
}

func (u *SimpleUpstream) GetRequest(vh *config.VirtualHost, pathStr string) (*http.Response, error) {
	newUrl, err := u.createNewUrl(vh, pathStr)
	if err != nil {
		return nil, err
	}

	return u.client.Get(newUrl.String())
}

func (u *SimpleUpstream) GetRequestPartial(vh *config.VirtualHost,
	pathStr string, start, end uint64) (*http.Response, error) {
	newUrl, err := u.createNewUrl(vh, pathStr)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("GET", newUrl.String(), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Range", fmt.Sprintf("bytes=%d-%d", start, end))
	return u.client.Do(req)
}

func (u *SimpleUpstream) Head(vh *config.VirtualHost, pathStr string) (*http.Response, error) {
	newUrl, err := u.createNewUrl(vh, pathStr)
	if err != nil {
		return nil, err
	}
	return u.client.Head(newUrl.String())
}

func (u *SimpleUpstream) GetSize(vh *config.VirtualHost, pathStr string) (int64, error) {
	resp, err := u.Head(vh, pathStr)
	if err != nil {
		return 0, err
	}

	return resp.ContentLength, nil
}

func (u *SimpleUpstream) GetHeader(vh *config.VirtualHost, pathStr string) (http.Header, error) {
	resp, err := u.Head(vh, pathStr)
	if err != nil {
		return nil, err
	}

	return resp.Header, nil
}

func (u *SimpleUpstream) createNewUrl(vh *config.VirtualHost, pathStr string) (*url.URL, error) {
	path, err := url.Parse(pathStr)
	if err != nil {
		return nil, err
	}

	return vh.UpstreamUrl().ResolveReference(path), nil
}
