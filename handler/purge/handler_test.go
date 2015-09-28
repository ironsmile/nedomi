package purge

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"golang.org/x/net/context"

	"github.com/ironsmile/nedomi/cache"
	"github.com/ironsmile/nedomi/config"
	"github.com/ironsmile/nedomi/contexts"
	"github.com/ironsmile/nedomi/logger"
	"github.com/ironsmile/nedomi/storage"
	"github.com/ironsmile/nedomi/types"
)

const (
	cacheKey    = "testkey"
	url         = "example.com/more/path*!:@#>"
	requestText = `
	{
		"cache_zone": "testZoen",
		"cache_zone_key": "testkey",
		"objects": [
			"/path/to/object",
			"/path/to/no/object"
		]
	}
	`
	badRequestText = `Bad request`
)

var obj1 = &types.ObjectMetadata{
	ID: types.NewObjectID(cacheKey, "/path/to/object"),
}

var obj2 = &types.ObjectMetadata{
	ID: types.NewObjectID(cacheKey, "/path/to/no/object"),
}

func removeFunctionMock(t *testing.T) func(parts ...*types.ObjectIndex) {
	return func(parts ...*types.ObjectIndex) {
		for _, part := range parts {
			if part.ObjID.CacheKey() != cacheKey {
				t.Fatalf("expected cache_key '%s' got '%s'", cacheKey, part.ObjID.CacheKey())
			}
			switch part.ObjID.Path() {
			case obj1.ID.Path(), obj2.ID.Path():
			default:
				t.Errorf("Remove was called with part with unexpected path %s", part.ObjID.Path())

			}
		}
	}
}

func testSetup(t *testing.T) (context.Context, *Handler, *types.Location) {
	var st = storage.NewMock(10)
	st.SaveMetadata(obj1)
	st.SavePart(
		&types.ObjectIndex{ObjID: obj1.ID, Part: 2},
		bytes.NewReader([]byte("test bytes")))
	st.SavePart(
		&types.ObjectIndex{ObjID: obj1.ID, Part: 4},
		bytes.NewReader([]byte("more bytes")))
	var cacheZoneMap = map[string]types.CacheZone{
		"testZoen": {
			Algorithm: cache.NewMock(&cache.MockRepliers{
				Remove: removeFunctionMock(t),
			}),
			Storage: st,
		},
	}
	loc := &types.Location{
		Logger: logger.NewMock(),
	}
	ctx := contexts.NewCacheZonesContext(context.Background(), cacheZoneMap)
	purger, err := New(&config.Handler{}, loc, nil)
	if err != nil {
		t.Fatal(err)
	}
	return ctx, purger, loc

}

func TestPurge(t *testing.T) {
	ctx, purger, loc := testSetup(t)
	req, err := http.NewRequest("POST", url,
		bytes.NewReader([]byte(requestText)))
	if err != nil {
		t.Fatal(err)
	}
	rec := httptest.NewRecorder()
	purger.RequestHandle(ctx, rec, req, loc)
	var pr purgeResult
	if err = json.Unmarshal(rec.Body.Bytes(), &pr); err != nil {
		t.Error(rec.Body.String())
		t.Fatal(err)
	}
	for key, value := range pr.Results {
		switch key {
		case obj1.ID.Path():
			if value != true {
				t.Errorf("result should've been true for path '%s'", key)
			}
		case obj2.ID.Path():
			if value != false {
				t.Errorf("result should've been false for path  '%s'", key)
			}
		default:
			t.Errorf("unxpected '%s':'%t' in the purge results", key, value)
		}
	}

}

func TestGetMethod(t *testing.T) {
	ctx, purger, loc := testSetup(t)
	req, err := http.NewRequest("GET", url,
		bytes.NewReader([]byte(requestText)))
	if err != nil {
		t.Fatal(err)
	}
	rec := httptest.NewRecorder()

	purger.RequestHandle(ctx, rec, req, loc)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("wrong response status %d expected %d",
			http.StatusMethodNotAllowed, rec.Code)
	}
}

func TestBadRequest(t *testing.T) {
	ctx, purger, loc := testSetup(t)
	req, err := http.NewRequest("POST", url,
		bytes.NewReader([]byte(badRequestText)))
	if err != nil {
		t.Fatal(err)
	}
	rec := httptest.NewRecorder()

	purger.RequestHandle(ctx, rec, req, loc)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("wrong response status %d expected %d",
			http.StatusBadRequest, rec.Code)
	}
}
