package upstream

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ironsmile/nedomi/types"
)

func testResponse(t *testing.T, u types.Upstream, path, expRespBody string, expRespCode int) {
	req, err := http.NewRequest("GET", "http://example.com"+path, nil)
	if err != nil {
		t.Fatal(err)
	}

	resp := httptest.NewRecorder()
	u.ServeHTTP(resp, req)

	if resp.Code != expRespCode {
		t.Errorf("Expected response code %d for %s but received %d", expRespCode, path, resp.Code)
	}
	if resp.Body.String() != expRespBody {
		t.Errorf("Expected response body %s for %s but received %s", expRespBody, path, resp.Body.String())
	}
}

func TestMockUpstream(t *testing.T) {
	t.Parallel()
	byeHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Bye")
	})
	errHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(400)
		fmt.Fprintf(w, "Error")
	})

	defaultUpstream := NewMock(nil)
	testResponse(t, defaultUpstream, "/test/", "Hello", 200)
	testResponse(t, defaultUpstream, "/error/", "Hello", 200)
	defaultUpstream.Handle("/error/", errHandler)
	testResponse(t, defaultUpstream, "/error/", "Error", 400)

	byeUpstream := NewMock(byeHandler)
	testResponse(t, byeUpstream, "/test/", "Bye", 200)
	testResponse(t, byeUpstream, "/error/", "Bye", 200)
	byeUpstream.Handle("/error/", errHandler)
	testResponse(t, byeUpstream, "/error/", "Error", 400)

}
