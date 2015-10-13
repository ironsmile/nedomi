package purge

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"golang.org/x/net/context"

	"github.com/ironsmile/nedomi/config"
	"github.com/ironsmile/nedomi/contexts"
	"github.com/ironsmile/nedomi/mock"
	"github.com/ironsmile/nedomi/types"
	"github.com/ironsmile/nedomi/utils/testutils"
)

const (
	cacheKey1, cacheKey2 = "testkey1", "testkey2"
	testURL              = "http://example.com/more/path*!:@#>"
	host1, host2, host3  = "example.org:1232", "example.net:8080", "notexample.net:8080"
	path1, path2, path3  = "/path/to/object", "/path/to/an/object", "/path/to/no/object"
	url1, url2, url3     = "http://" + host1 + path1, "http://" + host2 + path2, "http://" + host3 + path3
	requestText          = `[ "` + url1 + `", "` + url2 + `", "` + url3 + `" ]`
	badRequestText       = `Bad request`
)

var (
	obj1 = types.NewObjectID(cacheKey1, path1)
	obj2 = types.NewObjectID(cacheKey2, path2)
)

func removeFunctionMock(t *testing.T) func(parts ...*types.ObjectIndex) {
	return func(parts ...*types.ObjectIndex) {
		for _, part := range parts {
			switch part.ObjID.Path() {
			case obj1.Path(), obj2.Path():
			default:
				t.Errorf("Remove was called with part with unexpected path %s", part.ObjID.Path())
			}
		}
	}
}

func storageWithObjects(t *testing.T, objs ...*types.ObjectID) types.Storage {
	var st = mock.NewStorage(10)
	for _, obj := range objs {
		testutils.ShouldntFail(t,
			st.SaveMetadata(&types.ObjectMetadata{ID: obj}),
			st.SavePart(
				&types.ObjectIndex{ObjID: obj, Part: 2},
				bytes.NewReader([]byte("test bytes"))),
			st.SavePart(
				&types.ObjectIndex{ObjID: obj, Part: 4},
				bytes.NewReader([]byte("more bytes"))),
		)
	}
	return st
}

type mockApp struct {
	types.App
	getLocationFor func(string, string) *types.Location
}

func (m *mockApp) GetLocationFor(host, path string) *types.Location {
	return m.getLocationFor(host, path)
}

func testSetup(t *testing.T) (context.Context, *Handler, *types.Location) {
	var cz = &types.CacheZone{
		ID: "testZoen",
		Algorithm: mock.NewCacheAlgorithm(&mock.CacheAlgorithmRepliers{
			Remove: removeFunctionMock(t),
		}),
		Storage: storageWithObjects(t, obj1, obj2),
	}
	loc1 := &types.Location{
		Logger:   mock.NewLogger(),
		Cache:    cz,
		CacheKey: cacheKey1,
		Name:     "location1",
	}
	loc2 := &types.Location{
		Logger:   mock.NewLogger(),
		Cache:    cz,
		CacheKey: cacheKey2,
		Name:     "location2",
	}
	app := &mockApp{
		getLocationFor: func(host, path string) *types.Location {
			if host == host1 && path == path1 {
				return loc1
			}
			if host == host2 && path == path2 {
				return loc2
			}

			return nil
		},
	}
	loc3 := &types.Location{
		Logger: mock.NewLogger(),
	}

	ctx := contexts.NewAppContext(context.Background(), app)
	purger, err := New(&config.Handler{}, loc3, nil)
	if err != nil {
		t.Fatal(err)
	}
	return ctx, purger, loc3
}

func TestPurge(t *testing.T) {
	ctx, purger, loc := testSetup(t)
	if mockLogger, ok := loc.Logger.(*mock.Logger); ok {
		defer func() {
			//!TODO implement logger that wraps testing.TB
			if !t.Failed() {
				return
			}

			for _, log := range mockLogger.Logged() {
				t.Log(log)
			}
		}()
	}
	req, err := http.NewRequest("POST", testURL,
		bytes.NewReader([]byte(requestText)))
	if err != nil {
		t.Fatal(err)
	}
	rec := httptest.NewRecorder()
	purger.RequestHandle(ctx, rec, req)
	var pr purgeResult
	if err = json.Unmarshal(rec.Body.Bytes(), &pr); err != nil {
		t.Error(rec.Body.String())
		t.Fatal(err)
	}
	for key, value := range pr {
		switch key {
		case url1, url2:
			if value != true {
				t.Errorf("result should've been true for path '%s'", key)
			}
		case url3:
			if value != false {
				t.Errorf("result should've been false for path  '%s'", key)
			}
		default:
			t.Errorf("unxpected '%s':'%t' in the purge results", key, value)
		}
	}
}

func TestGetMethod(t *testing.T) {
	ctx, purger, _ := testSetup(t)
	req, err := http.NewRequest("GET", testURL,
		bytes.NewReader([]byte(requestText)))
	if err != nil {
		t.Fatal(err)
	}
	rec := httptest.NewRecorder()

	purger.RequestHandle(ctx, rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("wrong response status %d expected %d",
			http.StatusMethodNotAllowed, rec.Code)
	}
}

func TestBadRequest(t *testing.T) {
	ctx, purger, _ := testSetup(t)
	req, err := http.NewRequest("POST", testURL,
		bytes.NewReader([]byte(badRequestText)))
	if err != nil {
		t.Fatal(err)
	}
	rec := httptest.NewRecorder()

	purger.RequestHandle(ctx, rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("wrong response status %d expected %d",
			http.StatusBadRequest, rec.Code)
	}
}
