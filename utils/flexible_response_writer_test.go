package utils_test

import (
	"bytes"
	"fmt"
	"net/http"
	"testing"

	"github.com/ironsmile/nedomi/types"
	"github.com/ironsmile/nedomi/upstream"
	"github.com/ironsmile/nedomi/utils"
)

func testHandler(t *testing.T, u types.Upstream, path, expRespBody string, expRespCode int) {
	req, err := http.NewRequest("GET", "http://example.com"+path, nil)
	if err != nil {
		t.Fatal(err)
	}

	hooked := false
	buf := new(bytes.Buffer)
	hook := func(frw *utils.FlexibleResponseWriter) {
		hooked = true
		if frw.Code != expRespCode {
			t.Errorf("Expected response code %d for %s but received %d", expRespCode, path, frw.Code)
		}
		frw.BodyWriter = utils.NopCloser(buf)
	}

	resp := utils.NewFlexibleResponseWriter(hook)
	u.ServeHTTP(resp, req)

	if !hooked {
		t.Errorf("The hook function did not execute")
	}
	if buf.String() != expRespBody {
		t.Errorf("Expected response body %s for %s but received %s", expRespBody, path, buf.String())
	}
}

func TestFlexibleResponseWriter(t *testing.T) {
	t.Parallel()
	u := upstream.NewMock(nil)

	testHandler(t, u, "/test/", upstream.MockDefaultResponse, upstream.MockDefaultResponseCode)
	testHandler(t, u, "/error/", upstream.MockDefaultResponse, upstream.MockDefaultResponseCode)

	u.Handle("/error/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(400)
		fmt.Fprintf(w, "Error")
	}))
	testHandler(t, u, "/error/", "Error", 400)
}

func TestExpectedWriteError(t *testing.T) {
	t.Parallel()
	noop := func(frw *utils.FlexibleResponseWriter) {}
	resp := utils.NewFlexibleResponseWriter(noop)

	if _, err := resp.Write([]byte("test")); err == nil {
		t.Errorf("Expected to receive error with no writer")
	}
}

func TestCloseEmptyFleixbleResponseWriter(t *testing.T) {
	t.Parallel()
	noop := func(frw *utils.FlexibleResponseWriter) {}
	resp := utils.NewFlexibleResponseWriter(noop)

	if err := resp.Close(); err != nil {
		t.Errorf("Expected to not receive error on closing with no writer")
	}
}
