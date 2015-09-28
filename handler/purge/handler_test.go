package purge

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"golang.org/x/net/context"

	"github.com/ironsmile/nedomi/cache"
	"github.com/ironsmile/nedomi/config"
	"github.com/ironsmile/nedomi/contexts"
	"github.com/ironsmile/nedomi/logger"
	"github.com/ironsmile/nedomi/types"
)

type storageDiscardMock struct {
	types.Storage
	discard func(*types.ObjectID) error
}

func (s *storageDiscardMock) Discard(oid *types.ObjectID) error {
	return s.discard(oid)
}

func testSetup(t *testing.T) (context.Context, *Handler, *types.Location) {
	var cacheKey = "test_key_of_caches"
	var cacheZoneMap = map[string]types.CacheZone{
		"testZoen": {
			Algorithm: cache.NewMock(&cache.MockRepliers{
				RemoveObject: func(oid *types.ObjectID) bool {
					if oid.CacheKey() != cacheKey {
						t.Fatalf("expected cache_key '%s' got '%s'", cacheKey, oid.CacheKey())
					}
					switch oid.Path() {
					case "/path/to/object":
						return true
					case "/path/to/no/object":
						return false
					default:
						t.Errorf("RemoveObject was called with unexpected path %s", oid.Path())

					}
					return false
				},
			}),
			Storage: &storageDiscardMock{
				discard: func(oid *types.ObjectID) error {
					if oid.CacheKey() != cacheKey {
						t.Fatalf("expected cache_key '%s' got '%s'", cacheKey, oid.CacheKey())
					}
					switch oid.Path() {
					case "/path/to/object":
						return nil
					case "/path/to/no/object":
						return os.ErrNotExist
					default:
						t.Errorf("RemoveObject was called with unexpected path %s", oid.Path())
						return os.ErrNotExist
					}

				},
			},
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
	req, err := http.NewRequest("POST", "example.com/more/path*!:@#>", bytes.NewReader([]byte(`
	{
		"cache_zone": "testZoen",
		"cache_zone_key": "test_key_of_caches",
		"objects": [
			"/path/to/object",
			"/path/to/no/object"
		]
	}
	`)))
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
		case "/path/to/object":
			if value != true {
				t.Errorf("it was expected for paths '%s' the result to be true but it was not", key)
			}
		case "/path/to/no/object":
			if value != false {
				t.Errorf("it was expected for paths '%s' the result to be false but it was not", key)
			}
		default:
			t.Errorf("unxpected path '%s' with result '%t' in the purge results", key, value)
		}
	}

}

func TestGetMethod(t *testing.T) {
	ctx, purger, loc := testSetup(t)
	req, err := http.NewRequest("GET", "example.com/more/path*!:@#>", bytes.NewReader([]byte(`
	{
		"cache_zone": "testZoen",
		"cache_zone_key": "test_key_of_caches",
		"objects": [
			"/path/to/object",
			"/path/too/objects*",
			"/path/too/no/objects*",
			"/path/to/no/object"
		]
	}
	`)))
	if err != nil {
		t.Fatal(err)
	}
	rec := httptest.NewRecorder()

	purger.RequestHandle(ctx, rec, req, loc)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("get request didn't not return status %d but %d", http.StatusMethodNotAllowed, rec.Code)
	}
}

func TestBadRequest(t *testing.T) {
	ctx, purger, loc := testSetup(t)
	req, err := http.NewRequest("POST", "example.com/more/path*!:@#>", bytes.NewReader([]byte(`
	{
		"cache_zone": "testZoen" BAD!!!! JSON!!!
		"cache_zone_key": "test_key_of_caches",
		"objects": [
			"/path/to/object",
			"/path/to/no/object"
		]
	}
	`)))
	if err != nil {
		t.Fatal(err)
	}
	rec := httptest.NewRecorder()

	purger.RequestHandle(ctx, rec, req, loc)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("get request didn't not return status %d but %d", http.StatusBadRequest, rec.Code)
	}
}
