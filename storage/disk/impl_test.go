package disk

import (
	"bytes"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/ironsmile/nedomi/config"
	"github.com/ironsmile/nedomi/logger"
	"github.com/ironsmile/nedomi/types"
)

var obj1 = &types.ObjectMetadata{
	ID:                types.NewObjectID("testkey", "/lorem/ipsum"),
	ResponseTimestamp: time.Now().Unix(),
	Headers:           http.Header{"test": []string{"mest"}},
}
var obj2 = &types.ObjectMetadata{
	ID:                types.NewObjectID("concern", "/doge?so=scare&very_parameters"),
	ResponseTimestamp: time.Now().Unix(),
	Headers:           http.Header{"how-to": []string{"header"}},
}
var obj3 = &types.ObjectMetadata{
	ID:                types.NewObjectID("concern", "/very/space**"),
	ResponseTimestamp: time.Now().Unix(),
	Headers:           http.Header{"so": []string{"galaxy", "amaze"}},
}

func checkFile(t *testing.T, d *Disk, filePath, expectedContents string) {
	if stat, err := os.Stat(filePath); err != nil {
		t.Errorf("Could not stat file %s: %s", filePath, err)
	} else if stat.Mode() != d.filePermissions {
		t.Errorf("File %s has wrong permissions", filePath)
	} else if stat.Size() == 0 {
		t.Errorf("File %s is empty", filePath)
	}

	if contents, err := ioutil.ReadFile(filePath); err != nil {
		t.Errorf("Could not read %s: %s", filePath, err)
	} else if !strings.Contains(string(contents), expectedContents) {
		t.Errorf("File %s does not contain %s", filePath, expectedContents)
	}
}

func saveMetadata(t *testing.T, d *Disk, obj *types.ObjectMetadata) {
	if err := d.SaveMetadata(obj); err != nil {
		t.Fatalf("Could not save metadata for %s: %s", obj.ID, err)
	}
	checkFile(t, d, d.getObjectMetadataPath(obj.ID), obj.ID.CacheKey())

	if read, err := d.GetMetadata(obj.ID); err != nil || read == nil {
		t.Errorf("Received unexpected error while getting metadata: %s", err)
	} else if !reflect.DeepEqual(*read, *obj) {
		t.Fatalf("Original and read objects differ: '%#v', '%#v'", obj, read)
	}
}

func savePart(t *testing.T, d *Disk, idx *types.ObjectIndex, contents string) {
	if err := d.SavePart(idx, strings.NewReader(contents)); err != nil {
		t.Fatalf("Could not save file part %s: %s", idx, err)
	}
	checkFile(t, d, d.getObjectIndexPath(idx), contents)

	if partReader, err := d.GetPart(idx); err != nil {
		t.Errorf("Received unexpected error while getting part: %s", err)
	} else if readContents, err := ioutil.ReadAll(partReader); err != nil {
		t.Errorf("Could not read saved part: %s", err)
	} else if string(readContents) != contents {
		t.Errorf("Expected the contents to be %s but read %s", contents, readContents)
	}
}

func TestBasicOperations(t *testing.T) {
	t.Parallel()
	d, _, cleanup := getTestDiskStorage(t, 10)
	defer cleanup()

	idx := &types.ObjectIndex{ObjID: obj3.ID, Part: 5}
	idxContents := "0123456789"

	if _, err := d.GetMetadata(obj3.ID); err == nil {
		t.Error("There should have been no such metadata")
	} else if !os.IsNotExist(err) {
		t.Errorf("The error should have been os.ErrNotExist, but it's %#v", err)
	}

	if _, err := d.GetPart(idx); err == nil {
		t.Error("There should have been no such part")
	} else if !os.IsNotExist(err) {
		t.Errorf("The error should have been os.ErrNotExist, but it's %#v", err)
	}

	saveMetadata(t, d, obj3)

	if err := d.SavePart(idx, strings.NewReader(idxContents+"extra")); err == nil {
		t.Error("Saving a bigger part file should fail")
	}

	savePart(t, d, idx, idxContents)

	// Test discarding
	if err := d.DiscardPart(idx); err != nil {
		t.Errorf("Received unexpected error while discarding part: %s", err)
	}
	if err := d.Discard(obj3.ID); err != nil {
		t.Errorf("Received unexpected error while discarding object: %s", err)
	}
	iteratorTester(t, d, iterResMap{}) // Test that there is nothing left
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

	saveMetadata(t, d, obj1)
	savePart(t, d, &types.ObjectIndex{ObjID: obj1.ID, Part: 3}, "0123456789")

	expRes1 := iterResVal{*obj1, types.ObjectIndexMap{3: struct{}{}}, true}
	expectedResults[*obj1.ID] = &expRes1
	iteratorTester(t, d, expectedResults)

	saveMetadata(t, d, obj2)
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

	saveMetadata(t, d, obj1)
	saveMetadata(t, d, obj3)

	if err := os.Rename(d.getObjectMetadataPath(obj1.ID), d.getObjectMetadataPath(obj3.ID)); err != nil {
		t.Fatalf("Received an unexpected error while mobing the object: %s", err)
	}

	if err := d.Iterate(callback); err == nil {
		t.Error("Expected to receive an error")
	}
	ioutil.WriteFile(d.getObjectMetadataPath(obj3.ID), []byte("wrong json!"), d.filePermissions)
	if err := d.Iterate(callback); err == nil {
		t.Error("Expected to receive an error")
	}

	os.RemoveAll(d.getObjectIDPath(obj3.ID))
	if err := d.Iterate(callback); err == nil {
		t.Error("Expected to receive an error")
	}

	os.RemoveAll(d.getObjectIDPath(obj1.ID))
	if err := d.Iterate(callback); err != nil {
		t.Errorf("Received an unexpected error: %s", err)
	}
}

func TestConcurrentSaves(t *testing.T) {
	t.Parallel()

	partSize := 7
	contents := "1234567"
	idx := &types.ObjectIndex{ObjID: obj2.ID, Part: 5}

	d, _, cleanup := getTestDiskStorage(t, partSize)
	defer cleanup()

	wg := sync.WaitGroup{}
	concurrentSaves := 10 + rand.Intn(20)
	for i := 0; i <= concurrentSaves; i++ {
		r, w := io.Pipe()
		wg.Add(1)
		go func(reader io.Reader) {
			time.Sleep(time.Duration(rand.Intn(250)) * time.Millisecond)
			saveMetadata(t, d, obj2)
			time.Sleep(time.Duration(rand.Intn(250)) * time.Millisecond)
			if err := d.SavePart(idx, reader); err != nil {
				t.Fatalf("Unexpected error while saving part: %s", err)
			}
			wg.Done()
		}(r)

		go func(writer io.WriteCloser) {
			for i := 0; i < partSize; i++ {
				writer.Write([]byte{contents[i]})
				time.Sleep(time.Duration(rand.Intn(150)) * time.Millisecond)
			}
			if err := writer.Close(); err != nil {
				t.Fatalf("A really unexpected error: %s", err)
			}
		}(w)
	}
	wg.Wait()

	result := &bytes.Buffer{}
	partReader, err := d.GetPart(idx)
	if err != nil {
		t.Errorf("Unexpected error while saving part: %s", err)
	} else if copied, err := io.Copy(result, partReader); err != nil {
		t.Errorf("Unexpected error while reading part: %s", err)
	} else if copied != int64(partSize) || result.String() != contents {
		t.Errorf("Read only %d byes. Got %s and expected %s", copied, result, contents)
	}

	iteratorTester(t, d, iterResMap{*obj2.ID: &iterResVal{*obj2, types.ObjectIndexMap{5: struct{}{}}, true}})
}

func TestConstructor(t *testing.T) {
	t.Parallel()
	workingDiskPath, cleanup := getTestFolder(t)
	defer cleanup()

	cfg := &config.CacheZone{Path: workingDiskPath, PartSize: 10}
	l := logger.NewMock()

	if _, err := New(nil, l); err == nil {
		t.Error("Expected to receive error with nil config")
	}
	if _, err := New(cfg, nil); err == nil {
		t.Error("Expected to receive error with nil logger")
	}

	if _, err := New(&config.CacheZone{Path: "/an/invalid/path", PartSize: 10}, l); err == nil {
		t.Error("Expected to receive error with an invalid path")
	}
	if _, err := New(&config.CacheZone{Path: "/", PartSize: 10}, l); err == nil {
		t.Error("Expected to receive error with root path")
	}
	if _, err := New(&config.CacheZone{Path: workingDiskPath, PartSize: 0}, l); err == nil {
		t.Error("Expected to receive error with invalid part size")
	}

	if _, err := New(cfg, l); err != nil {
		t.Errorf("Received unexpected error while creating a normal disk storage: %s", err)
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


*/
