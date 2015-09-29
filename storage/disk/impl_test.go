package disk

import (
	"bytes"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"reflect"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/ironsmile/nedomi/config"
	"github.com/ironsmile/nedomi/logger"
	"github.com/ironsmile/nedomi/types"
	"github.com/ironsmile/nedomi/utils/testutils"
)

func init() {
	rand.Seed(time.Now().Unix())
}

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

func randSleep(fromMs, toMs int) {
	time.Sleep(time.Duration(fromMs+rand.Intn(toMs)) * time.Millisecond)
}

func saveMetadata(t *testing.T, d *Disk, obj *types.ObjectMetadata) {
	if err := d.SaveMetadata(obj); err != nil {
		t.Fatalf("Could not save metadata for %s: %s", obj.ID, err)
	}
	checkFile(t, d, d.getObjectMetadataPath(obj.ID), obj.ID.CacheKey())
	randSleep(0, 50)
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
	if parts, err := d.GetAvailableParts(idx.ObjID); err != nil {
		t.Errorf("Received unexpected error while getting available parts: %s", err)
	} else if _, ok := parts[idx.Part]; !ok {
		t.Errorf("Could not find the saved part %s", idx)
	}

	randSleep(0, 50)
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
	if parts, err := d.GetAvailableParts(obj3.ID); err != nil {
		t.Errorf("Received unexpected error while getting available parts: %s", err)
	} else if len(parts) > 0 {
		t.Errorf("Should not have got parts but received %#v", parts)
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
	cpus := runtime.NumCPU()
	runtime.GOMAXPROCS(cpus)

	partSize := 7
	contents := "1234567"
	idx := &types.ObjectIndex{ObjID: obj2.ID, Part: 5}

	d, _, cleanup := getTestDiskStorage(t, partSize)
	defer cleanup()

	wg := sync.WaitGroup{}
	concurrentSaves := 10 + rand.Intn(cpus*3)
	for i := 0; i <= concurrentSaves; i++ {
		r, w := io.Pipe()
		wg.Add(1)
		go func(reader io.Reader) {
			randSleep(0, 150)
			saveMetadata(t, d, obj2)
			randSleep(0, 150)
			if err := d.SavePart(idx, reader); err != nil {
				t.Fatalf("Unexpected error while saving part: %s", err)
			}
			wg.Done()
		}(r)

		go func(writer io.WriteCloser) {
			for i := 0; i < partSize; i++ {
				writer.Write([]byte{contents[i]})
				randSleep(0, 100)
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
	workingDiskPath, cleanup := testutils.GetTestFolder(t)
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
