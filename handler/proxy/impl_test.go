package proxy

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"testing"

	"github.com/ironsmile/nedomi/config"
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
	upstream, err := New(&config.Handler{}, &types.Location{
		Name:     "test",
		Logger:   mock.NewLogger(),
		Upstream: upstream.NewSimple(upstreamURL),
	}, nil)
	if err != nil {
		t.Fatal(err)
	}

	req1, err := http.NewRequest("GET", "http://www.somewhere.com/err", nil)
	if err != nil {
		t.Fatal(err)
	}
	resp1 := httptest.NewRecorder()
	upstream.ServeHTTP(resp1, req1)
	if resp1.Code != 404 || resp1.Body.String() != "error!" {
		t.Errorf("Unexpected response %#v", resp1)
	}

	req2, err := http.NewRequest("GET", "http://www.somewhere.com/index", nil)
	if err != nil {
		t.Fatal(err)
	}
	resp2 := httptest.NewRecorder()
	upstream.ServeHTTP(resp2, req2)
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
	upstream, err := New(&config.Handler{}, &types.Location{
		Name:     "test",
		Logger:   mock.NewLogger(),
		Upstream: upstream.NewSimple(upstreamURL),
	}, nil)
	if err != nil {
		t.Fatal(err)
	}

	resp := httptest.NewRecorder()
	upstream.ServeHTTP(resp, req)

	if !responded {
		t.Errorf("Server did not respond")
	}
	if resp.Body.String() != "boo" {
		t.Errorf("Unexpected response %s", resp.Body)
	}
}
