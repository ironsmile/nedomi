package flv

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
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
	return types.RequestHandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request, l *types.Location) {
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
	v := setup(t)
	var req = makeRequest(t, "/test.flv")
	var rec = httptest.NewRecorder()
	v.RequestHandle(nil, rec, req, nil)
	var got, expected = rec.Body.String(), fsmap["test.flv"]
	if got != expected {
		t.Errorf("flv handler: didn't return file without a start parameter, expected `%s` got `%s`", expected, got)
	}
}

func TestFlvWithParam(t *testing.T) {
	v := setup(t)
	var req = makeRequest(t, "/test.flv?start=20")
	var rec = httptest.NewRecorder()
	v.RequestHandle(nil, rec, req, nil)
	var expected, got = fsmap["test.flv"][20:], rec.Body.String()[13:]
	if got != expected {
		t.Errorf("flv handler: didn't return file from the correct position with start parameter, expected `%s` got `%s`", expected, got)
	}
	checkHeader(t, rec.Body.Bytes())
}

func TestFlv404(t *testing.T) {
	v := setup(t)
	var req = makeRequest(t, "/nonexistant?start=2040")
	var rec = httptest.NewRecorder()
	v.RequestHandle(nil, rec, req, nil)
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
	if !reflect.DeepEqual(content[:13], flvHeader[:]) {
		t.Errorf("header is not correct\nExpected :`%#v`\nGot      :`%#v`", flvHeader[:], content[:13])
	}

}
