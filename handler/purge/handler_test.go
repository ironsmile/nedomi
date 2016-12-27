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
	host4, host5, host6  = "example.org:1232", "example.org:notPOrt", "notexample.net:8080"
	path4, path5, path6  = "/path/to/no/object", "/path/to/object", "/path/to/no/object"
	url4, url5, url6     = "http://" + host4 + path4, "http://" + host5 + path5, "http://" + host6 + path6
	requestText          = `[ "` + url1 + `", "` + url2 + `", "` + url3 + `", "` + url4 + `", "` + url5 + `", "` + url6 + `" ]`
	badRequestText       = `Bad request`
)

var (
	obj1 = types.NewObjectID(cacheKey1, path1)
	obj2 = types.NewObjectID(cacheKey2, path2)
)

func testCode(t *testing.T, code, expected int) {
	if code != expected {
		t.Fatalf("wrong response code %d expected %d", code, expected)
	}
}

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

func storageWithObjectWithGetPartsErrors(t *testing.T, objs ...*types.ObjectID) types.Storage {
	storage := storageWithObjects(t, objs...)
	return newBadGetParts(t, storage, objs[len(objs)-1])
}

func storageWithObjectWithDiscardErrors(t *testing.T, objs ...*types.ObjectID) types.Storage {
	storage := storageWithObjects(t, objs...)
	return newBadDiscard(t, storage, objs[len(objs)-1])
}

type mockApp struct {
	types.App
	getLocationFor func(string, string) *types.Location
}

func (m *mockApp) GetLocationFor(host, path string) *types.Location {
	return m.getLocationFor(host, path)
}

func testSetup(t *testing.T) (context.Context, *Handler, *types.Location) {
	return testSetupWithStorage(t, storageWithObjects(t, obj1, obj2))
}

func testSetupWithStorage(t *testing.T, st types.Storage) (context.Context, *Handler, *types.Location) {
	var cz = &types.CacheZone{
		ID: "testZoen",
		Algorithm: mock.NewCacheAlgorithm(&mock.CacheAlgorithmRepliers{
			Remove: removeFunctionMock(t),
		}),
		Storage: st,
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
			if host == host1 {
				return loc1
			}
			if host == host2 {
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
	purger.ServeHTTP(ctx, rec, req)
	testCode(t, rec.Code, http.StatusOK)
	var pr purgeResult
	if err = json.Unmarshal(rec.Body.Bytes(), &pr); err != nil {
		t.Error(rec.Body.String())
		t.Fatal(err)
	}
	var (
		toBeTrue  = []string{url1, url2}
		toBeFalse = []string{url3, url4, url5, url6}
	)

	checkPr(t, pr, toBeTrue, true)
	checkPr(t, pr, toBeFalse, false)
	if len(pr) > len(toBeTrue)+len(toBeFalse) {
		t.Errorf("have %d purges in the result instead of %d true ones and %d false ones:\n %+v",
			len(pr), len(toBeTrue), len(toBeFalse), pr)
	}
}

func checkPr(t *testing.T, pr purgeResult, urls []string, expected bool) {
	for _, key := range urls {
		if value, ok := pr[key]; !ok {
			t.Errorf("expected the pr to have a key %s but it didn't", key)
		} else if value != expected {
			t.Errorf("result should've been '%t' for path '%s'", expected, key)
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

	purger.ServeHTTP(ctx, rec, req)
	testCode(t, rec.Code, http.StatusMethodNotAllowed)
}

func TestBadRequest(t *testing.T) {
	ctx, purger, _ := testSetup(t)
	req, err := http.NewRequest("POST", testURL,
		bytes.NewReader([]byte(badRequestText)))
	if err != nil {
		t.Fatal(err)
	}
	rec := httptest.NewRecorder()

	purger.ServeHTTP(ctx, rec, req)
	testCode(t, rec.Code, http.StatusBadRequest)
}

func TestNoApp(t *testing.T) {
	_, purger, _ := testSetup(t)
	req, err := http.NewRequest("POST", testURL,
		bytes.NewReader([]byte(requestText)))
	if err != nil {
		t.Fatal(err)
	}
	rec := httptest.NewRecorder()

	purger.ServeHTTP(context.Background(), rec, req)
	testCode(t, rec.Code, http.StatusInternalServerError)
}

func TestPurgeBadDiscard(t *testing.T) {
	ctx, purger, loc := testSetupWithStorage(
		t, storageWithObjectWithDiscardErrors(t, obj1, obj2))
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
	purger.ServeHTTP(ctx, rec, req)
	testCode(t, rec.Code, http.StatusInternalServerError)
}

func TestPurgeBadGetParts(t *testing.T) {
	ctx, purger, loc := testSetupWithStorage(
		t, storageWithObjectWithGetPartsErrors(t, obj1, obj2))
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
	purger.ServeHTTP(ctx, rec, req)
	testCode(t, rec.Code, http.StatusInternalServerError)
}
