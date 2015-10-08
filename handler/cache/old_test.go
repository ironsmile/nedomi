package cache

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"runtime"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/ironsmile/nedomi/cache"
	"github.com/ironsmile/nedomi/config"
	"github.com/ironsmile/nedomi/handler/mock"
	"github.com/ironsmile/nedomi/logger"
	"github.com/ironsmile/nedomi/storage"
	"github.com/ironsmile/nedomi/types"
	"github.com/ironsmile/nedomi/utils/httputils"
	"github.com/ironsmile/nedomi/utils/testutils"

	"golang.org/x/net/context"

	"golang.org/x/tools/godoc/vfs/httpfs"
	"golang.org/x/tools/godoc/vfs/mapfs"
)

func init() {
	rand.Seed(time.Now().Unix())
}

var fsmap = map[string]string{
	"test.flv": "This is FLV test data. As there is noting that requires the data to be actual valid flv a strings is fine.",
	"path":     "awesome test path content of the test material more letters/bytes to be tested with",
}

func fsMapHandler() http.HandlerFunc {
	return func(wr http.ResponseWriter, req *http.Request) {
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

func setup(t *testing.T) (*mock.Handler, *types.Location, config.CacheZone, int, func()) {
	cpus := runtime.NumCPU()
	goroutines := cpus * 4
	runtime.GOMAXPROCS(cpus * 20)
	up := mock.NewHandler(fsMapHandler())
	loc := &types.Location{}
	var err error
	loc.Logger = newStdLogger()

	path, cleanup := testutils.GetTestFolder(t)

	cz := config.CacheZone{
		ID:             "1",
		Type:           "disk",
		Path:           path,
		StorageObjects: 200000,
		PartSize:       5,
	}

	ca := cache.NewMock(&cache.MockRepliers{
		Lookup:     func(*types.ObjectIndex) bool { return true },
		ShouldKeep: func(*types.ObjectIndex) bool { return true },
	})
	st, err := storage.New(&cz, loc.Logger)
	if err != nil {
		panic(err)
	}
	loc.CacheKey = "test"
	loc.Cache = &types.CacheZone{
		ID:        cz.ID,
		PartSize:  cz.PartSize,
		Algorithm: ca,
		Scheduler: storage.NewScheduler(),
		Storage:   st,
	}

	return up, loc, cz, goroutines, cleanup
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
	t.Parallel()
	up, loc, _, goroutines, cleanup := setup(t)
	defer cleanup()

	headerKeyFunc := func(i int) string {
		return fmt.Sprintf("X-Header-%d", i)
	}

	headerValueFunc := func(i int) string {
		return fmt.Sprintf("value-%d", i)
	}

	// setup the response
	for i := 0; i < 100; i++ {
		handler := func(num int) func(w http.ResponseWriter, r *http.Request) {
			return func(w http.ResponseWriter, r *http.Request) {
				w.Header().Add(headerKeyFunc(num), headerValueFunc(num))
				w.WriteHeader(200)
				fmt.Fprintf(w, fsmap["path"])
			}
		}(i)

		up.Handle("/path/"+strconv.Itoa(i), http.HandlerFunc(handler))
	}

	cacheHandler, err := New(nil, loc, up)
	if err != nil {
		t.Fatal(err)
	}

	testFunc := func(t *testing.T, i, j int) {
		rec := httptest.NewRecorder()
		req, err := http.NewRequest("HEAD", "/path/"+strconv.Itoa(i), nil)
		if err != nil {
			t.Fatal(err)
		}
		cacheHandler.RequestHandle(context.Background(), rec, req)
		val := rec.Header().Get(headerKeyFunc(i))
		expVal := headerValueFunc(i)
		if val != expVal {
			t.Errorf("Expected header value %s and received %s", expVal, val)
		}
	}
	concurrentTestHelper(t, goroutines, 100, testFunc)
}

func TestStorageSimultaneousGets(t *testing.T) {
	t.Parallel()
	expected := fsmap["path"]
	up, loc, _, goroutines, cleanup := setup(t)
	defer cleanup()
	runtime.GOMAXPROCS(runtime.NumCPU())
	cacheHandler, err := New(nil, loc, up)
	if err != nil {
		t.Fatal(err)
	}
	concurrentTestHelper(t, goroutines, 100, func(t *testing.T, i, j int) {
		rec := httptest.NewRecorder()
		req, err := http.NewRequest("GET", "http://example.com/path", nil)
		if err != nil {
			t.Fatal(err)
		}
		cacheHandler.RequestHandle(context.Background(), rec, req)
		if rec.Code != http.StatusOK {
			t.Errorf("Got code different from OK %d", rec.Code)
		}
		b, err := ioutil.ReadAll(rec.Body)
		if err != nil {
			t.Errorf("Got error while reading response on %d, %d: %s", j, i, err)
		}
		if string(b) != expected {
			t.Errorf("The response was expected to be \n'%s'\n but it was \n'%s'", expected, string(b))
		}
	})
}

func TestStorageSimultaneousRangeGets(t *testing.T) {
	t.Parallel()
	var expected = fsmap["path"]
	up, loc, _, goroutines, cleanup := setup(t)
	defer cleanup()
	runtime.GOMAXPROCS(runtime.NumCPU())
	cacheHandler, err := New(nil, loc, up)
	if err != nil {
		t.Fatal(err)
	}
	testfunc := func(t *testing.T, i, j int) {
		var begin = rand.Intn(len(expected) - 4)
		var length = rand.Intn(len(expected)-begin-1) + 2
		var rec = httptest.NewRecorder()
		ran := httputils.Range{
			Start:  uint64(begin),
			Length: uint64(length),
		}
		req, err := http.NewRequest("GET", "http://example.com/path", nil)
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Add("Range", ran.Range())
		cacheHandler.RequestHandle(context.Background(), rec, req)
		if rec.Code != http.StatusPartialContent {
			t.Errorf("Got code different from OK %d", rec.Code)
		}
		b, err := ioutil.ReadAll(rec.Body)
		if err != nil {
			t.Errorf("Got error while reading response on %d, %d: %s", j, i, err)
		}
		expected := expected[begin : begin+length]
		if string(b) != expected {
			t.Errorf("The response for `%+v`was expected to be \n'%s'\n but it was \n'%s'", req, expected, string(b))
		}
	}

	concurrentTestHelper(t, goroutines, 100, testfunc)
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
