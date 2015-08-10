package disk

import (
	"fmt"
	"net/http"
	"net/url"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ironsmile/nedomi/config"
	"github.com/ironsmile/nedomi/types"
	"github.com/ironsmile/nedomi/upstream"
)

// Tests the storage headers map in multithreading usage. An error will be
// triggered by a race condition. This may or may not happen. So no failures
// in the test do not mean there are no problems. But it may catch an error
// from time to time. In fact it does quite often for me.
//
// Most of the time the test fails with a panic. And most of the time
// the panic is in the runtime. So isntead of a error message via t.Error
// the test fails with a panic.
func TestStorageHeadersFunctionWithManuGoroutines(t *testing.T) {
	runtime.GOMAXPROCS(runtime.NumCPU())

	httpSrv := &http.Server{
		Addr:    "127.0.0.1:54231",
		Handler: &testHandler{},
	}

	//!TODO: cleanup this webserver
	go httpSrv.ListenAndServe()

	cz := config.CacheZoneSection{}
	cz.CacheAlgo = "not-an-algo"
	cz.PartSize = 1024
	cz.Path = "./test"
	cz.StorageObjects = 1024

	cm := &cacheManagerMock{}

	// The upstream is not actually using the passed config structure.
	// For the moment it is safe to give it anything.
	// Or maybe it would be better to create a mock upstream for testing.
	URL, err := url.Parse("http://127.0.0.1:54231")
	if err != nil {
		t.Fatal(err)
	}

	up, err := upstream.New("simple", URL)

	if err != nil {
		t.Fatalf("Test upstream was not ceated. %s", err)
	}

	vh := &config.VirtualHost{}
	vh.UpstreamAddress = URL.String()
	vh.Verify(make(map[uint32]*config.CacheZoneSection))

	storage := New(cz, cm, up)

	var wg sync.WaitGroup

	for i := 0; i < 16; i++ {
		wg.Add(1)

		go func(j int) {
			defer wg.Done()

			// A vain attempt to catch the panic. Most of the times
			// it is in the runtime goroutine, though.
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Panic in goroutine %d. %s", j, r)
				}
			}()

			for i := 0; i < 100; i++ {
				oid := types.ObjectID{}
				oid.CacheKey = "1"
				oid.Path = fmt.Sprintf("/%d", i)

				_, err := storage.Headers(oid)
				if err != nil {
					t.Errorf("Got error from storage.Headers on %d, %d: %s", j, i, err)
				}
			}
		}(i)
	}

	wg.Wait()
}

type countingUpstream struct {
	*fakeUpstream
	called int32
}

func (c *countingUpstream) GetRequestPartial(path string, start, end uint64) (*http.Response, error) {
	atomic.AddInt32(&c.called, 1)
	return c.fakeUpstream.GetRequestPartial(path, start, end)
}

func TestStorageSimultaneousGets(t *testing.T) {
	runtime.GOMAXPROCS(runtime.NumCPU())

	cz := config.CacheZoneSection{}
	cz.CacheAlgo = "not-an-algo"
	cz.PartSize = 1024
	cz.Path = "./test"
	cz.StorageObjects = 1024

	cm := &cacheManagerMock{}

	up := &countingUpstream{
		fakeUpstream: NewFakeUpstream(),
	}

	up.addFakeResponse("path",
		FakeResponse{
			Status:       "200",
			ResponseTime: 20 * time.Nanosecond,
			Response:     "awesome",
		})

	storage := New(cz, cm, up)

	var wg sync.WaitGroup

	for i := 0; i < 16; i++ {
		wg.Add(1)

		go func(j int) {
			defer wg.Done()

			oid := types.ObjectID{}
			oid.CacheKey = "1"
			oid.Path = "path"
			_, err := storage.GetFullFile(oid)
			if err != nil {
				t.Errorf("Got error from storage.Get on %d, %d: %s", j, i, err)
			}
		}(i)
	}

	wg.Wait()
	if up.called != 1 {
		t.Errorf("Expected upstream.GetRequest to be called once it got called %d", up.called)
	}
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
