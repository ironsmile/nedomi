package proxy

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"testing"

	"golang.org/x/net/context"

	"github.com/ironsmile/nedomi/config"
	"github.com/ironsmile/nedomi/contexts"
	"github.com/ironsmile/nedomi/mock"
	"github.com/ironsmile/nedomi/types"
	"github.com/ironsmile/nedomi/upstream"
)

func TestSimpleUpstream(t *testing.T) {
	t.Parallel()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/err" {
			w.WriteHeader(404)
			fmt.Fprint(w, "error!")
			return
		}
		w.WriteHeader(200)
		fmt.Fprint(w, "hello world")
	}))
	defer ts.Close()

	upstreamURL, err := url.Parse(ts.URL)
	if err != nil {
		t.Fatal(err)
	}
	up, err := upstream.NewSimple(upstreamURL)
	if err != nil {
		t.Fatal(err)
	}

	proxy, err := New(&config.Handler{}, &types.Location{
		Name:     "test",
		Logger:   mock.NewLogger(),
		Upstream: up,
	}, nil)
	if err != nil {
		t.Fatal(err)
	}

	req1, err := http.NewRequest("GET", "http://www.somewhere.com/err", nil)
	if err != nil {
		t.Fatal(err)
	}
	resp1 := httptest.NewRecorder()
	proxy.ServeHTTP(context.Background(), resp1, req1)
	if resp1.Code != 404 || resp1.Body.String() != "error!" {
		t.Errorf("Unexpected response %#v", resp1)
	}

	req2, err := http.NewRequest("GET", "http://www.somewhere.com/index", nil)
	if err != nil {
		t.Fatal(err)
	}
	resp2 := httptest.NewRecorder()
	proxy.ServeHTTP(context.Background(), resp2, req2)
	if resp2.Code != 200 || resp2.Body.String() != "hello world" {
		t.Errorf("Unexpected response %#v", resp2)
	}

}

func TestSimpleUpstreamHeaders(t *testing.T) {
	t.Parallel()

	req, err := http.NewRequest("GET", "http://www.somewhere.com/err", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("test", "mest")
	req.Header.Set("User-Agent", "nedomi") // The only exception...
	headersCopy := make(http.Header)
	for k, v := range req.Header {
		headersCopy[k] = v
	}

	responded := false
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !reflect.DeepEqual(headersCopy, r.Header) {
			t.Errorf("Different request headers: expected %#v, received %#v", headersCopy, r.Header)
		}
		responded = true
		fmt.Fprint(w, "boo")
	}))
	defer ts.Close()

	upstreamURL, err := url.Parse(ts.URL)
	if err != nil {
		t.Fatal(err)
	}
	up, err := upstream.NewSimple(upstreamURL)
	if err != nil {
		t.Fatal(err)
	}

	proxy, err := New(&config.Handler{}, &types.Location{
		Name:     "test",
		Logger:   mock.NewLogger(),
		Upstream: up,
	}, nil)
	if err != nil {
		t.Fatal(err)
	}

	resp := httptest.NewRecorder()
	proxy.ServeHTTP(context.Background(), resp, req)

	if !responded {
		t.Errorf("Server did not respond")
	}
	if resp.Body.String() != "boo" {
		t.Errorf("Unexpected response %s", resp.Body)
	}
}

func TestSimpleRetry(t *testing.T) {
	t.Parallel()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/err" {
			w.WriteHeader(404)
			fmt.Fprint(w, "error!")
			return
		}
		w.WriteHeader(200)
		fmt.Fprint(w, "hello world")
	}))
	defer ts.Close()

	upstreamURL, err := url.Parse(ts.URL)
	if err != nil {
		t.Fatal(err)
	}
	up, err := upstream.NewSimple(upstreamURL)
	if err != nil {
		t.Fatal(err)
	}

	retryTs := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/err" {
			w.WriteHeader(200)
			fmt.Fprint(w, "not error!")
			return
		}
		w.WriteHeader(404)
		fmt.Fprint(w, "not hello world")
	}))
	defer retryTs.Close()

	retryUpstreamURL, err := url.Parse(retryTs.URL)
	if err != nil {
		t.Fatal(err)
	}
	retryUpstream, err := upstream.NewSimple(retryUpstreamURL)
	if err != nil {
		t.Fatal(err)
	}

	proxy, err := New(
		config.NewHandler("proxy", json.RawMessage(`{ "try_other_upstream_on_code" : {"404": "retry_upstream"}}`)),
		&types.Location{
			Name:     "test",
			Logger:   mock.NewLogger(),
			Upstream: up,
		}, nil)
	if err != nil {
		t.Fatal(err)
	}

	req1, err := http.NewRequest("GET", "http://www.somewhere.com/err", nil)
	if err != nil {
		t.Fatal(err)
	}
	resp1 := httptest.NewRecorder()
	ctx := contexts.NewAppContext(context.Background(), &mockApp{
		upstreams: map[string]types.Upstream{
			"retry_upstream": retryUpstream,
		},
	})
	proxy.ServeHTTP(ctx, resp1, req1)
	if resp1.Code != 200 || resp1.Body.String() != "not error!" {
		t.Errorf("Unexpected response %#v", resp1)
	}

	req2, err := http.NewRequest("GET", "http://www.somewhere.com/index", nil)
	if err != nil {
		t.Fatal(err)
	}
	resp2 := httptest.NewRecorder()
	proxy.ServeHTTP(context.Background(), resp2, req2)
	if resp2.Code != 200 || resp2.Body.String() != "hello world" {
		t.Errorf("Unexpected response %#v", resp2)
	}
}

type mockApp struct {
	types.App
	upstreams map[string]types.Upstream
}

func (m *mockApp) GetUpstream(id string) types.Upstream {
	return m.upstreams[id]
}

func TestSimpleRetryWithNilUpstream(t *testing.T) {
	t.Parallel()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/err" {
			w.WriteHeader(404)
			fmt.Fprint(w, "error!")
			return
		}
		w.WriteHeader(200)
		fmt.Fprint(w, "hello world")
	}))
	defer ts.Close()

	upstreamURL, err := url.Parse(ts.URL)
	if err != nil {
		t.Fatal(err)
	}
	up, err := upstream.NewSimple(upstreamURL)
	if err != nil {
		t.Fatal(err)
	}

	retryTs := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/err" {
			w.WriteHeader(200)
			fmt.Fprint(w, "not error!")
			return
		}
		w.WriteHeader(404)
		fmt.Fprint(w, "not hello world")
	}))
	defer retryTs.Close()

	retryUpstreamURL, err := url.Parse(retryTs.URL)
	if err != nil {
		t.Fatal(err)
	}
	retryUpstream, err := upstream.NewSimple(retryUpstreamURL)
	if err != nil {
		t.Fatal(err)
	}

	proxy, err := New(
		config.NewHandler("proxy", json.RawMessage(`{ "try_other_upstream_on_code" : {"404": "nonexistant_upstream"}}`)),
		&types.Location{
			Name:     "test",
			Logger:   mock.NewLogger(),
			Upstream: up,
		}, nil)
	if err != nil {
		t.Fatal(err)
	}

	req1, err := http.NewRequest("GET", "http://www.somewhere.com/err", nil)
	if err != nil {
		t.Fatal(err)
	}
	resp1 := httptest.NewRecorder()
	ctx := contexts.NewAppContext(context.Background(), &mockApp{
		upstreams: map[string]types.Upstream{
			"retry_upstream": retryUpstream,
		},
	})
	proxy.ServeHTTP(ctx, resp1, req1)
	if resp1.Code != 404 || resp1.Body.String() != "error!" {
		t.Errorf("Unexpected response %#v", resp1)
	}

	req2, err := http.NewRequest("GET", "http://www.somewhere.com/index", nil)
	if err != nil {
		t.Fatal(err)
	}
	resp2 := httptest.NewRecorder()
	proxy.ServeHTTP(context.Background(), resp2, req2)
	if resp2.Code != 200 || resp2.Body.String() != "hello world" {
		t.Errorf("Unexpected response %#v", resp2)
	}
}
