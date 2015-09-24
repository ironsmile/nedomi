package cache

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"os"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/ironsmile/nedomi/cache"
	"github.com/ironsmile/nedomi/config"
	"github.com/ironsmile/nedomi/contexts"
	"github.com/ironsmile/nedomi/logger"
	"github.com/ironsmile/nedomi/storage"
	"github.com/ironsmile/nedomi/types"
	"github.com/ironsmile/nedomi/upstream"
	"github.com/ironsmile/nedomi/utils"

	"golang.org/x/net/context"

	"golang.org/x/tools/godoc/vfs/httpfs"
	"golang.org/x/tools/godoc/vfs/mapfs"
)

var fsmap = map[string]string{
	"test.flv": "This is FLV test data. As there is noting that requires the data to be actual valid flv a strings is fine.",
	"path":     "awesome",
}

func fsMapHandler() http.HandlerFunc {
	return func(wr http.ResponseWriter, req *http.Request) {
		time.Sleep(time.Millisecond) // long enough
		wr.Header().Add("Expires", time.Now().Add(time.Hour).Format(time.RFC1123))
		http.FileServer(httpfs.New(mapfs.New(fsmap))).ServeHTTP(wr, req)
	}
}

func newStdLogger() types.Logger {
	l, err := logger.New(config.NewLogger("std", []byte(`{"level":"error"}`)))
	if err != nil {
		panic(err)
	}
	return l
}

func setup() (types.Upstream, *types.Location, config.CacheZone, int) {
	cpus := runtime.NumCPU()
	goroutines := cpus * 4
	runtime.GOMAXPROCS(cpus)
	up := upstream.NewMock(fsMapHandler())
	loc := &types.Location{Upstream: up}
	loc.Upstream = up
	var err error
	loc.Logger = newStdLogger()

	path := "/tmp/cachestorage"
	os.RemoveAll(path)
	os.MkdirAll(path, 0777)

	cz := config.CacheZone{
		ID:             "1",
		Type:           "disk",
		Path:           path,
		StorageObjects: 200000,
		PartSize:       2,
	}

	ca := cache.NewMock(&cache.MockReplies{
		Lookup:     true,
		ShouldKeep: true,
		AddObject:  nil,
	})
	st, err := storage.New(&cz, loc.Logger)
	if err != nil {
		panic(err)
	}
	loc.CacheKey = cz.ID
	loc.Cache.Storage = st
	loc.Cache.Algorithm = ca

	return up, loc, cz, goroutines
}

// Tests the storage headers map in multithreading usage. An error will be
// triggered by a race condition. This may or may not happen. So no failures
// in the test do not mean there are no problems. But it may catch an error
// from time to time. In fact it does quite often for me.
//
// Most of the time the test fails with a panic. And most of the time
// the panic is in the runtime. So isntead of a error message via t.Error
// the test fails with a panic.
func TestStorageHeadersFunctionWithManyGoroutines(t *testing.T) {
	t.SkipNow()
	_, loc, _, goroutines := setup()

	headerKeyFunc := func(i int) string {
		return fmt.Sprintf("X-Header-%d", i)
	}

	headerValueFunc := func(i int) string {
		return fmt.Sprintf("value-%d", i)
	}

	// setup the response
	for i := 0; i < 100; i++ {
		var headers = make(http.Header)
		headers.Add(headerKeyFunc(i), headerValueFunc(i))
	}
	ctx := contexts.NewLocationContext(context.Background(), loc)

	cacheHandler, err := New(nil, loc, nil)
	if err != nil {
		t.Fatal(err)
	}

	testFunc := func(t *testing.T, i, j int) {
		fmt.Println("started", i, j)
		rec := httptest.NewRecorder()
		req, err := http.NewRequest("HEAD", "/path", nil)
		if err != nil {
			t.Fatal(err)
		}
		cacheHandler.RequestHandle(ctx, rec, req, nil)
		fmt.Println("finished", i, j)
		fmt.Println("finished", rec.Header())
		_ = rec.Header().Get(headerKeyFunc(i))
		if t.Failed() {
			t.FailNow()
		}
	}
	concurrentTestHelper(t, goroutines, 100, testFunc)
}

func TestStorageSimultaneousGets(t *testing.T) {
	up, loc, _, goroutines := setup()
	runtime.GOMAXPROCS(runtime.NumCPU())
	ctx := contexts.NewLocationContext(context.Background(), &types.Location{Upstream: up})
	cacheHandler, err := New(nil, loc, nil)
	if err != nil {
		t.Fatal(err)
	}
	concurrentTestHelper(t, goroutines, 1, func(t *testing.T, i, j int) {
		rec := httptest.NewRecorder()
		req, err := http.NewRequest("GET", "http://example.com/path", nil)
		if err != nil {
			t.Fatal(err)
		}
		cacheHandler.RequestHandle(ctx, rec, req, nil)
		if rec.Code != http.StatusOK {
			t.Errorf("Got code different from OK %d", rec.Code)
		}
		b, err := ioutil.ReadAll(rec.Body)
		if err != nil {
			t.Errorf("Got error while reading response on %d, %d: %s", j, i, err)
		}
		if string(b) != "awesome" {
			t.Errorf("The response was expected to be 'awesome' but it was %s", string(b))
		}

		if t.Failed() {
			t.FailNow()
		}
	})
}

func TestStorageSimultaneousRangeGets(t *testing.T) {
	var expected = fsmap["path"]
	up, loc, _, goroutines := setup()
	runtime.GOMAXPROCS(runtime.NumCPU())
	ctx := contexts.NewLocationContext(context.Background(), &types.Location{Upstream: up})
	cacheHandler, err := New(nil, loc, nil)
	if err != nil {
		t.Fatal(err)
	}
	testfunc := func(t *testing.T, i, j int) {
		var begin = rand.Intn(len(expected) - 4)
		var length = rand.Intn(len(expected)-begin-1) + 2
		var rec = httptest.NewRecorder()
		ran := utils.HTTPRange{
			Start:  uint64(begin),
			Length: uint64(length),
		}
		req, err := http.NewRequest("GET", "http://example.com/path", nil)
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Add("Range", ran.Range())
		cacheHandler.RequestHandle(ctx, rec, req, nil)
		if rec.Code != http.StatusPartialContent {
			t.Errorf("Got code different from OK %d", rec.Code)
		}
		b, err := ioutil.ReadAll(rec.Body)
		if err != nil {
			t.Errorf("Got error while reading response on %d, %d: %s", j, i, err)
		}
		expected := expected[begin : begin+length]
		if string(b) != expected {
			t.Errorf("The response for `%+v`was expected to be '%s' but it was %s", req, expected, string(b))
		}

		if t.Failed() {
			t.FailNow()
		}
	}

	concurrentTestHelper(t, goroutines, 1, testfunc)
}

func concurrentTestHelper(t *testing.T, goroutines, iterations int, test func(t *testing.T, i, j int)) {
	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func(j int) {
			defer wg.Done()
			for i := 0; i < iterations; i++ {
				test(t, i, j)
			}
		}(i)
	}

	wg.Wait()
}
