package disk

import (
	"fmt"
	"net/http"
	"net/url"
	"runtime"
	"sync"
	"testing"

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
	cz.Algorithm = "not-an-algo"
	cz.PartSize = 1024
	cz.Path = "/some/path"
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

	vh := &config.VirtualHost{UpstreamAddress: URL}
	if err := vh.Validate(); err != nil {
		t.Fatal(err)
	}

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
				oid.Path = fmt.Sprintf("/%d/%d", i, j)

				storage.Headers(oid)
			}
		}(i)
	}

	wg.Wait()
}
