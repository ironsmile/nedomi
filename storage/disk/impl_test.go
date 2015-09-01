package disk

import (
	"io/ioutil"
	"net/http"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/ironsmile/nedomi/types"
)

//!TODO: move this to a utils file?
func wait(t *testing.T, period time.Duration, errorMessage string, action func()) {
	finished := make(chan struct{})

	go func() {
		action()
		close(finished)
	}()

	select {
	case <-finished:
	case <-time.After(period):
		t.Errorf("Test exceeded allowed time of %d seconds: %s", period/time.Second, errorMessage)
	}
	return
}

var obj1 = &types.ObjectMetadata{
	ID: &types.ObjectID{
		CacheKey: "testkey",
		Path:     "/lorem/ipsum",
	},
	ResponseTime: time.Now(),
	Size:         121,
	Headers:      http.Header{"test": []string{"mest"}},
	IsCacheable:  true,
}
var obj2 = &types.ObjectMetadata{
	ID: &types.ObjectID{
		CacheKey: "concern",
		Path:     "/doge?so=scare&very_parameters",
	},
	ResponseTime: time.Now(),
	Size:         50,
	Headers:      http.Header{"how-to": []string{"header"}},
	IsCacheable:  true,
}
var obj3 = &types.ObjectMetadata{
	ID: &types.ObjectID{
		CacheKey: "concern",
		Path:     "/very/space**",
	},
	Size:         7,
	ResponseTime: time.Now(),
	Headers:      http.Header{"so": []string{"galaxy", "amaze"}},
	IsCacheable:  true,
}

func TestBasicOperations(t *testing.T) {
	t.Parallel()
	t.Skip("TODO: test saving, loading and discarding of metadata and parts")
}

type iterResVal struct {
	obj            types.ObjectMetadata
	parts          types.ObjectIndexMap
	ShouldContinue bool
}
type iterResMap map[types.ObjectID]*iterResVal

func getCallback(t *testing.T, expectedResults iterResMap) (func(*types.ObjectMetadata, types.ObjectIndexMap) bool, *int) {
	receivedResults := 0
	localExpResults := iterResMap{}
	for k, v := range expectedResults {
		localExpResults[k] = v
	}
	// Every test error is fatal to prevent endless loops
	return func(obj *types.ObjectMetadata, parts types.ObjectIndexMap) bool {
		if obj == nil || parts == nil {
			t.Fatal("Received a nil result")
		}
		if len(localExpResults) == 0 {
			t.Fatalf("Received a result '%#v' when there are no expected results", obj)
		}
		expRes, ok := localExpResults[*obj.ID]
		if !ok {
			t.Fatalf("Received an unexpected result '%#v'", obj)
		}
		if !reflect.DeepEqual(*obj, expRes.obj) {
			t.Fatalf("Received and expected objects differ: '%#v', '%#v'", *obj, expRes.obj)
		}
		if !reflect.DeepEqual(parts, expRes.parts) {
			t.Fatalf("Received and expected parts differ: '%#v', '%#v'", parts, expRes.parts)
		}
		delete(localExpResults, *obj.ID) // We expect to receive each object only once
		receivedResults++

		return expRes.ShouldContinue
	}, &receivedResults
}

func iteratorTester(t *testing.T, d *Disk, expectedResults iterResMap) {
	expectedResultsNum := len(expectedResults)
	callback, resultsNum := getCallback(t, expectedResults)
	if err := d.Iterate(callback); err != nil {
		t.Errorf("Received an unexpected error when iterating: %s", err)
	}

	if *resultsNum != expectedResultsNum {
		t.Errorf("Expected %d results but received %d", expectedResultsNum, *resultsNum)
	}
}

func TestIteration(t *testing.T) {
	t.Parallel()
	d, _, cleanup := getTestDiskStorage(t, 10)
	defer cleanup()

	expectedResults := iterResMap{}
	iteratorTester(t, d, expectedResults)

	if err := d.SaveMetadata(obj1); err != nil {
		t.Fatalf("Could not save metadata for %s: %s", obj1.ID, err)
	}
	idx := &types.ObjectIndex{ObjID: obj1.ID, Part: 3}
	if err := d.SavePart(idx, strings.NewReader("0123456789")); err != nil {
		t.Fatalf("Could not save file part %s: %s", idx, err)
	}
	expRes1 := iterResVal{*obj1, types.ObjectIndexMap{3: struct{}{}}, true}
	expectedResults[*obj1.ID] = &expRes1
	iteratorTester(t, d, expectedResults)

	if err := d.SaveMetadata(obj2); err != nil {
		t.Fatalf("Could not save metadata for %s: %s", obj2.ID, err)
	}
	expRes2 := iterResVal{*obj2, types.ObjectIndexMap{}, true}
	expectedResults[*obj2.ID] = &expRes2
	iteratorTester(t, d, expectedResults)

	// Test stopping
	expRes1.ShouldContinue = false
	expRes2.ShouldContinue = false
	callback, resultsNum := getCallback(t, expectedResults)
	if err := d.Iterate(callback); err != nil {
		t.Errorf("Received an unexpected error when iterating: %s", err)
	}
	if *resultsNum != 1 {
		t.Errorf("Expected iteration to stop immediately, but instead got %d results", *resultsNum)
	}
}

func TestIterationErrors(t *testing.T) {
	t.Parallel()
	d, _, cleanup := getTestDiskStorage(t, 10)
	defer cleanup()

	callback, _ := getCallback(t, iterResMap{})

	if err := d.SaveMetadata(obj1); err != nil {
		t.Fatalf("Could not save metadata for %s: %s", obj1.ID, err)
	}
	if err := d.SaveMetadata(obj2); err != nil {
		t.Fatalf("Could not save metadata for %s: %s", obj2.ID, err)
	}

	if err := os.Rename(d.getObjectMetadataPath(obj1.ID), d.getObjectMetadataPath(obj2.ID)); err != nil {
		t.Fatalf("Received an unexpected error while mobing the object: %s", err)
	}

	if err := d.Iterate(callback); err == nil {
		t.Error("Expected to receive an error")
	}
	ioutil.WriteFile(d.getObjectMetadataPath(obj2.ID), []byte("wrong json!"), d.filePermissions)
	if err := d.Iterate(callback); err == nil {
		t.Error("Expected to receive an error")
	}

	os.RemoveAll(d.getObjectIDPath(obj2.ID))
	if err := d.Iterate(callback); err == nil {
		t.Error("Expected to receive an error")
	}

	os.RemoveAll(d.getObjectIDPath(obj1.ID))
	if err := d.Iterate(callback); err != nil {
		t.Errorf("Received an unexpected error: %s", err)
	}
}

/*
func setup() (*fakeUpstream, config.CacheZoneSection, *CacheAlgorithmMock, int) {
	cpus := runtime.NumCPU()
	goroutines := cpus * 4
	runtime.GOMAXPROCS(cpus)

	cz := config.CacheZoneSection{}
	cz.Algorithm = "not-an-algo"
	cz.PartSize = 1024
	cz.Path = os.TempDir() + "/nedomi-test" //!TODO: use random dirs; cleanup after use
	cz.StorageObjects = 1024

	ca := &CacheAlgorithmMock{}
	up := newFakeUpstream()
	return up, cz, ca, goroutines

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
	up, cz, ca, goroutines := setup()

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
			fakeResponse{
				Status:       "200",
				ResponseTime: 10 * time.Nanosecond,
				Headers:      headers,
			},
		)
	}
	storage := New(cz, ca, newStdLogger())
	defer storage.Close()
	defer os.RemoveAll(storage.path)
	ctx := contexts.NewVhostContext(context.Background(), &types.VirtualHost{Upstream: up})

	concurrentTestHelper(t, goroutines, 100, func(t *testing.T, i, j int) {
		oid := types.ObjectID{}
		oid.CacheKey = "1"
		oid.Path = pathFunc(i)

		header, err := storage.Headers(ctx, oid)
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
	fakeup, cz, ca, goroutines := setup()
	runtime.GOMAXPROCS(runtime.NumCPU())
	up := &countingUpstream{
		fakeUpstream: fakeup,
	}

	up.addFakeResponse("/path",
		fakeResponse{
			Status:       "200",
			ResponseTime: 10 * time.Millisecond,
			Response:     "awesome",
		})

	storage := New(cz, ca, newStdLogger())
	defer storage.Close()
	defer os.RemoveAll(storage.path)
	ctx := contexts.NewVhostContext(context.Background(), &types.VirtualHost{Upstream: up})

	concurrentTestHelper(t, goroutines, 1, func(t *testing.T, i, j int) {
		oid := types.ObjectID{}
		oid.CacheKey = "1"
		oid.Path = "/path"
		file, err := storage.GetFullFile(ctx, oid)
		if err != nil {
			t.Errorf("Got error from storage.Get on %d, %d: %s", j, i, err)
		}
		b, err := ioutil.ReadAll(file)
		if err != nil {
			t.Errorf("Got error while reading response on %d, %d: %s", j, i, err)
		}
		if string(b) != "awesome" {
			t.Errorf("The response was expected to be 'awesome' but it was %s", string(b))
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

		for resultIndex, value := range result {
			if value.Part != test.result[resultIndex] {
				t.Errorf("Wrong part for test index %d, wanted %d in position %d but got %d", index, test.result[resultIndex], resultIndex, value.Part)
			}
		}
	}
}

func newStdLogger() types.Logger {
	l, _ := logger.New(config.LoggerSection{
		Type:     "std",
		Settings: []byte(`{"level":"debug"}`),
	})
	return l
}
*/
