package simple

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestSimpleUpstream(t *testing.T) {
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
	upstream := New(upstreamURL)

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
