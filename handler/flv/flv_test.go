package flv

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strconv"
	"testing"

	"golang.org/x/net/context"

	"github.com/ironsmile/nedomi/config"
	"github.com/ironsmile/nedomi/types"
	"golang.org/x/tools/godoc/vfs/httpfs"
	"golang.org/x/tools/godoc/vfs/mapfs"
)

var fsmap = map[string]string{
	"test.flv": "This is FLV test data. As there is noting that requires the data to be actual valid flv a strings is fine.",
}

func fsMapHandler() types.RequestHandler {
	var fileHandler = http.FileServer(httpfs.New(mapfs.New(fsmap)))
	return types.RequestHandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		fileHandler.ServeHTTP(w, r)
	})

}

func setup(t *testing.T) types.RequestHandler {
	var v, err = New(config.NewHandler("flv", json.RawMessage(``)), nil, fsMapHandler())
	if err != nil {
		t.Fatalf("error on creating new flv handler - %s", err)
	}

	return v
}

func TestFlvNotParam(t *testing.T) {
	t.Parallel()
	v := setup(t)
	var req = makeRequest(t, "/test.flv")
	var rec = httptest.NewRecorder()
	v.RequestHandle(nil, rec, req)
	var got, expected = rec.Body.String(), fsmap["test.flv"]
	if got != expected {
		t.Errorf("flv handler: didn't return file without a start parameter, expected `%s` got `%s`", expected, got)
	}
}

func TestFlvWithParam(t *testing.T) {
	t.Parallel()
	v := setup(t)
	var start = 20
	var fileContent = fsmap["test.flv"]
	var expectedContentLength = len(fileContent) - start + len(flvHeader)
	var req = makeRequest(t, "/test.flv?start="+strconv.Itoa(start))
	var rec = httptest.NewRecorder()
	v.RequestHandle(nil, rec, req)
	var expected, got = fileContent[start:], rec.Body.String()[len(flvHeader):]
	if got != expected {
		t.Errorf("flv handler: didn't return file from the correct position with start parameter, expected `%s` got `%s`", expected, got)
	}
	checkHeader(t, rec.Body.Bytes())

	var contentLengthStr = rec.Header().Get("Content-Length")
	if contentLengthStr == "" {
		t.Errorf("not Content-Length but it was expected")
	}
	var contentLength, err = strconv.Atoi(contentLengthStr)
	if err != nil {
		t.Errorf("error parsing Content-Length : %s", err)
	}
	if contentLength != expectedContentLength {
		t.Errorf("Expected Content-Length to be %d got %d", expectedContentLength, contentLength)

	}
}

func TestFlv404(t *testing.T) {
	t.Parallel()
	v := setup(t)
	var req = makeRequest(t, "/nonexistant?start=2040")
	var rec = httptest.NewRecorder()
	v.RequestHandle(nil, rec, req)
	var expected, got = "404 page not found\n", rec.Body.String()
	if rec.Code != http.StatusNotFound {
		t.Errorf("flv handler: code not %d on not existant request but %d", http.StatusNotFound, rec.Code)
	}
	if got != expected {
		t.Errorf("flv handler: didn't return 404, expected `%s` got `%s`", expected, got)
	}
}

func makeRequest(t *testing.T, url string) *http.Request {
	var req, err = http.NewRequest("GET", url, nil)
	if err != nil {
		t.Fatalf("error while creating http request - %s", err)
	}

	return req
}

func checkHeader(t *testing.T, content []byte) {
	if !reflect.DeepEqual(content[:len(flvHeader)], flvHeader[:]) {
		t.Errorf("header is not correct\nExpected :`%#v`\nGot      :`%#v`", flvHeader[:], content[:len(flvHeader)])
	}

}

func TestRecalculateContentLenght(t *testing.T) {
	var tests = map[string]string{
		"12": "25", // normal
		"":   "",   // missing
		"a2": "",   // wat!?
	}
	for input, expected := range tests {
		var header = http.Header{}
		header.Set("Content-Length", input)
		recalculateContentLength(header)
		var got = header.Get("Content-Length")
		if got != expected {
			t.Errorf("got '%s', expected '%s'", got, expected)
		}
	}
}

func TestFlvWithWriteError(t *testing.T) { // for the coverage
	t.Parallel()
	v := setup(t)
	var start = 20
	var errAfter = len(flvHeader) / 2
	var req = makeRequest(t, "/test.flv?start="+strconv.Itoa(start))
	var expectedErr = fmt.Errorf("expected error")
	var rec = httptest.NewRecorder()
	v.RequestHandle(nil, newErrAtWriter(rec, errAfter, expectedErr), req)

	if rec.Body.Len() != errAfter {
		t.Errorf("didn't stop writing when error occured")
	}
}

type errAtWriter struct {
	http.ResponseWriter
	n   int
	err error
}

func newErrAtWriter(rw http.ResponseWriter, n int, err error) http.ResponseWriter {
	return &errAtWriter{
		ResponseWriter: rw,
		n:              n,
		err:            err,
	}
}

func (e *errAtWriter) Write(p []byte) (n int, err error) {
	var toBeWritten = min(len(p), e.n)
	if toBeWritten > 0 {
		n, err = e.ResponseWriter.Write(p[:toBeWritten])
	}
	e.n -= toBeWritten
	if err == nil && e.n == 0 {
		return n, e.err
	}
	return n, err
}

func min(l, r int) int {
	if l > r {
		return r
	}
	return l
}
