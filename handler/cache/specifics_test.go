package cache

import (
	"encoding/hex"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"runtime"
	"testing"
	"unsafe"

	"github.com/ironsmile/nedomi/cache"
	"github.com/ironsmile/nedomi/config"
	"github.com/ironsmile/nedomi/mock"
	"github.com/ironsmile/nedomi/storage"
	"github.com/ironsmile/nedomi/types"
	"github.com/ironsmile/nedomi/utils/httputils"
	"github.com/ironsmile/nedomi/utils/testutils"
	"golang.org/x/net/context"
)

type testApp struct {
	testing.TB
	cacheHandler *CachingProxy
	ctx          context.Context
}

func reqForRange(path string, begin, length uint64) *http.Request {
	ran := httputils.Range{Start: begin, Length: length}
	req, err := http.NewRequest("GET", "http://example.com/"+path, nil)
	if err != nil {
		panic(err)
	}

	req.Header.Add("Range", ran.Range())
	return req
}

func testRequest(t *testApp, req *http.Request, expected string, code int) {
	var rec = httptest.NewRecorder()
	t.cacheHandler.RequestHandle(t.ctx, rec, req)
	if rec.Code != code {
		t.Errorf("Got code different from %d - %d", code, rec.Code)
	}
	b, err := ioutil.ReadAll(rec.Body)
	if err != nil {
		t.Errorf("Got error while reading response on %+v: %s", req, err)
	}

	if string(b) != expected {
		t.Errorf("The response for `%+v`was expected to be \n'%s'\n but it was \n'%s'",
			req, expected, string(b))
	}
}

func testRange(t *testApp, path string, begin, length uint64) {
	expected := fsmap[path]
	req := reqForRange(path, begin, length)
	testRequest(t, req, expected[begin:begin+length], http.StatusPartialContent)
}

func testFullRequest(t *testApp, path string) {
	expected := fsmap[path]
	req, err := http.NewRequest("GET", "http://example.com/"+path, nil)
	if err != nil {
		panic(err)
	}

	testRequest(t, req, expected, http.StatusOK)
}

func TestVeryFragmentedFile(t *testing.T) {
	var file = "fragmented"
	fsmap[file] = generateMeAString(1, 1024)
	up, loc, _, _, cleanup := realerSetup(t)
	defer cleanup()
	cacheHandler, err := New(nil, loc, up)
	if err != nil {
		t.Fatal(err)
	}
	app := &testApp{
		TB:           t,
		ctx:          context.Background(),
		cacheHandler: cacheHandler,
	}
	testRange(app, file, 5, 10)
	testRange(app, file, 5, 2)
	testRange(app, file, 2, 2)
	testRange(app, file, 20, 10)
	testRange(app, file, 30, 10)
	testRange(app, file, 40, 10)
	testRange(app, file, 50, 10)
	testRange(app, file, 60, 10)
	testRange(app, file, 70, 10)
	testRange(app, file, 50, 20)
	testRange(app, file, 200, 5)
	testFullRequest(app, file)
	testRange(app, file, 3, 1000)
}

func realerSetup(t testing.TB) (*mock.RequestHandler, *types.Location, *config.CacheZone, int, func()) {
	cpus := runtime.NumCPU()
	goroutines := cpus * 4
	runtime.GOMAXPROCS(cpus)
	up := mock.NewRequestHandler(fsMapHandler())
	loc := &types.Location{}
	var err error
	loc.Logger = newStdLogger()
	loc.CacheKey = "test"
	loc.CacheKeyIncludesQuery = false

	path, cleanup := testutils.GetTestFolder(t)

	cz := &config.CacheZone{
		ID:             "1",
		Type:           "disk",
		Path:           path,
		StorageObjects: 20,
		Algorithm:      "lru",
		PartSize:       5,
	}

	st, err := storage.New(cz, loc.Logger)
	if err != nil {
		panic(err)
	}
	ca, err := cache.New(cz, st.DiscardPart, loc.Logger)
	if err != nil {
		panic(err)
	}
	loc.Cache = &types.CacheZone{
		ID:        cz.ID,
		PartSize:  cz.PartSize,
		Algorithm: ca,
		Scheduler: storage.NewScheduler(),
		Storage:   st,
	}

	return up, loc, cz, goroutines, cleanup
}

// generates a pseudo-randomized string to be used as the contents of a file
func generateMeAString(seed, size int64) string {
	var b = make([]byte, size)
	r := readerFromSource(rand.NewSource(seed))
	if _, err := r.Read(b); err != nil {
		panic(err)
	}

	return hex.EncodeToString(b)[:size]
}

type sourceReader struct {
	rand.Source
}

func (s *sourceReader) Read(b []byte) (int, error) {
	l := len(b)
	var tmp int64
	for a := 0; l > a; a += 8 {
		tmp = s.Int63()
		copy(b[a:], (*[8]byte)(unsafe.Pointer(&tmp))[:])

	}
	return l, nil
}

func readerFromSource(s rand.Source) io.Reader {
	return &sourceReader{s}
}

func Test2PartsFile(t *testing.T) {
	var file = "2parts"
	fsmap[file] = generateMeAString(2, 10)
	up, loc, _, _, cleanup := realerSetup(t)
	defer cleanup()
	cacheHandler, err := New(nil, loc, up)
	if err != nil {
		t.Fatal(err)
	}
	app := &testApp{
		TB:           t,
		ctx:          context.Background(),
		cacheHandler: cacheHandler,
	}
	testRange(app, file, 2, 8)
	testFullRequest(app, file)
}
