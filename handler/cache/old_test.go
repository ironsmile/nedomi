package cache

import (
	"fmt"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync"
	"testing"

	"github.com/ironsmile/nedomi/utils/testutils"
)

// Tests the storage headers map in multithreading usage. An error will be
// triggered by a race condition. This may or may not happen. So no failures
// in the test do not mean there are no problems. But it may catch an error
// from time to time. In fact it does quite often for me.
//
// Most of the time the test fails with a panic. And most of the time
// the panic is in the runtime. So instead of a error message via t.Error
// the test fails with a panic.
func TestStorageHeadersFunctionWithManyGoroutines(t *testing.T) {
	t.Parallel()
	app := newTestApp(t)
	defer app.cleanup()

	headerKeyFunc := func(i int) string {
		return fmt.Sprintf("X-Header-%d", i)
	}

	headerValueFunc := func(i int) string {
		return fmt.Sprintf("value-%d", i)
	}
	var pathFile = testutils.GenerateMeAString(20, 30)

	// setup the response
	for i := 0; i < 100; i++ {
		handler := func(num int) func(w http.ResponseWriter, r *http.Request) {
			return func(w http.ResponseWriter, r *http.Request) {
				w.Header().Add(headerKeyFunc(num), headerValueFunc(num))
				fmt.Fprintf(w, pathFile)
			}
		}(i)

		app.up.Handle("/path/"+strconv.Itoa(i), http.HandlerFunc(handler))
	}

	testFunc := func(t *testing.T, i, j int) {
		rec := httptest.NewRecorder()
		req, err := http.NewRequest("HEAD", "/path/"+strconv.Itoa(i), nil)
		if err != nil {
			t.Fatal(err)
		}
		app.cacheHandler.ServeHTTP(rec, req)
		val := rec.Header().Get(headerKeyFunc(i))
		expVal := headerValueFunc(i)
		if val != expVal {
			t.Errorf("Expected header value %s and received %s", expVal, val)
		}
	}
	concurrentTestHelper(t, 20, 100, testFunc)
}

func TestStorageSimultaneousGets(t *testing.T) {
	t.Parallel()
	app := newTestApp(t)
	defer app.cleanup()
	var files = app.getFileSizes()
	files = files[:len(files)/2] // it takes too long otherwise
	concurrentTestHelper(t, len(files)*5, 50, func(t *testing.T, i, j int) {
		app.testFullRequest(files[(i*j)%len(files)].path)
	})
}

func TestStorageSimultaneousRangeGets(t *testing.T) {
	t.Parallel()
	app := newTestApp(t)
	defer app.cleanup()
	var files = app.getFileSizes()
	testfunc := func(t *testing.T, i, j int) {
		var file = files[(i*j)%len(files)]
		var begin = rand.Intn(file.size - 4)
		var length = rand.Intn(file.size-begin-1) + 2
		app.testRange(file.path, uint64(begin), uint64(length))
	}

	concurrentTestHelper(t, len(files)*5, 50, testfunc)
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
