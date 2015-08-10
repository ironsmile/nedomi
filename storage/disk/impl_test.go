package disk

import (
	"fmt"
	"net/http"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/ironsmile/nedomi/config"
	"github.com/ironsmile/nedomi/logger"
	"github.com/ironsmile/nedomi/types"
)

func setup() (*fakeUpstream, config.CacheZoneSection, *cacheManagerMock, int) {
	cpus := runtime.NumCPU()
	goroutines := cpus * 4
	runtime.GOMAXPROCS(cpus)

	cz := config.CacheZoneSection{}
	cz.CacheAlgo = "not-an-algo"
	cz.PartSize = 1024
	cz.Path = "./test"
	cz.StorageObjects = 1024

	cm := &cacheManagerMock{}
	up := NewFakeUpstream()
	return up, cz, cm, goroutines

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
	up, cz, cm, goroutines := setup()

	pathFunc := func(i int) string {
		return fmt.Sprintf("/%d", i)
	}

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
		up.addFakeResponse(pathFunc(i),
			FakeResponse{
				Status:       "200",
				ResponseTime: 10 * time.Nanosecond,
				Headers:      headers,
			},
		)
	}
	storage := New(cz, cm, up, newStdLogger())

	concurrentTestHelper(t, goroutines, 100, func(t *testing.T, i, j int) {
		oid := types.ObjectID{}
		oid.CacheKey = "1"
		oid.Path = pathFunc(i)

		header, err := storage.Headers(oid)
		if err != nil {
			t.Errorf("Got error from storage.Headers on %d, %d: %s", j, i, err)
		}
		value := header.Get(headerKeyFunc(i))
		if value != headerValueFunc(i) {
			t.Errorf("Expected header [%s] to have value [%s] but it had value %s for %d, %d", headerKeyFunc(i), headerValueFunc(i), value, j, i)
		}
	})

}

func TestStorageSimultaneousGets(t *testing.T) {
	fakeup, cz, cm, goroutines := setup()
	runtime.GOMAXPROCS(runtime.NumCPU())
	up := &countingUpstream{
		fakeUpstream: fakeup,
	}

	up.addFakeResponse("path",
		FakeResponse{
			Status:       "200",
			ResponseTime: 20 * time.Nanosecond,
			Response:     "awesome",
		})

	storage := New(cz, cm, up, newStdLogger())

	concurrentTestHelper(t, goroutines, 1, func(t *testing.T, i, j int) {
		oid := types.ObjectID{}
		oid.CacheKey = "1"
		oid.Path = "path"
		_, err := storage.GetFullFile(oid)
		if err != nil {
			t.Errorf("Got error from storage.Get on %d, %d: %s", j, i, err)
		}
	})

	if up.called != 1 {
		t.Errorf("Expected upstream.GetRequest to be called once it got called %d", up.called)
	}
}

func concurrentTestHelper(t *testing.T, goroutines, iterations int, test func(t *testing.T, i, j int)) {
	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func(j int) {
			defer wg.Done()

			// A vain attempt to catch the panic. Most of the times
			// it is in the runtime goroutine, though.
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Panic in goroutine %d. %s", j, r)
				}
			}()

			for i := 0; i < iterations; i++ {
				test(t, i, j)
			}
		}(i)
	}

	wg.Wait()
}

var breakInIndexesMatrix = []struct {
	ID       types.ObjectID
	start    uint64
	end      uint64
	partSize uint64
	result   []uint32
}{{
	ID:       types.ObjectID{},
	start:    0,
	end:      99,
	partSize: 50,
	result:   []uint32{0, 1},
}, {
	ID:       types.ObjectID{},
	start:    5,
	end:      99,
	partSize: 50,
	result:   []uint32{0, 1},
}, {
	ID:       types.ObjectID{},
	start:    50,
	end:      99,
	partSize: 50,
	result:   []uint32{1},
}, {
	ID:       types.ObjectID{},
	start:    50,
	end:      50,
	partSize: 50,
	result:   []uint32{1},
}, {
	ID:       types.ObjectID{},
	start:    50,
	end:      49,
	partSize: 50,
	result:   []uint32{},
},
}

func TestBreakInIndexes(t *testing.T) {
	for index, test := range breakInIndexesMatrix {
		var result = breakInIndexes(test.ID, test.start, test.end, test.partSize)
		if len(result) != len(test.result) {
			t.Errorf("Wrong len (%d != %d) on test index %d", len(result), len(test.result), index)
		}

		for resultIndex, _ := range result {
			if result[resultIndex].Part != test.result[resultIndex] {
				t.Errorf("Wrong part for test index %d, wanted %d in position %d but got %d", index, test.result[resultIndex], resultIndex, result[resultIndex].Part)
			}
		}
	}
}

func newStdLogger() logger.Logger {
	l, _ := logger.New("std", config.LoggerSection{
		Type:     "std",
		Settings: []byte(`{"level":"debug"}`),
	})
	return l
}
