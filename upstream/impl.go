package upstream

import (
	"fmt"
	"net/http"
	"net/url"
)

type impl struct {
	client       http.Client
	upstreamHost *url.URL
}

func New(host string) (Upstream, error) {
	hostUrl, err := url.Parse(host)
	if err != nil {
		return nil, err
	}

	return &impl{
		upstreamHost: hostUrl,
		client:       http.Client{},
	}, nil
}

func (u *impl) GetRequest(pathStr string) (*http.Response, error) {

	newUrl, err := u.createNewUrl(pathStr)
	if err != nil {
		return nil, err
	}

	return u.client.Get(newUrl.String())
}

func (u *impl) GetRequestPartial(pathStr string, start, end uint64) (*http.Response, error) {

	newUrl, err := u.createNewUrl(pathStr)
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

func (u *impl) Head(pathStr string) (*http.Response, error) {
	newUrl, err := u.createNewUrl(pathStr)
	if err != nil {
		return nil, err
	}
	return u.client.Head(newUrl.String())
}

func (u *impl) GetSize(pathStr string) (int64, error) {
	resp, err := u.Head(pathStr)
	if err != nil {
		return 0, err
	}

	return resp.ContentLength, nil
}

func (u *impl) GetHeader(pathStr string) (http.Header, error) {
	resp, err := u.Head(pathStr)
	if err != nil {
		return nil, err
	}

	return resp.Header, nil
}

func (u *impl) createNewUrl(pathStr string) (*url.URL, error) {
	path, err := url.Parse(pathStr)
	if err != nil {
		return nil, err
	}

	return u.upstreamHost.ResolveReference(path), nil
}
