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

func (u *impl) GetRequest(pathStr string, start, end uint64) (*http.Response, error) {
	path, err := url.Parse(pathStr)
	if err != nil {
		return nil, err
	}
	newUrl := u.upstreamHost.ResolveReference(path)
	req, err := http.NewRequest("GET", newUrl.String(), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Range", fmt.Sprintf("bytes=%d-%d", start, end))
	return u.client.Do(req)
}
