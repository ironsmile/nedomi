package cache

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"runtime"
	"testing"

	"github.com/ironsmile/nedomi/cache"
	"github.com/ironsmile/nedomi/config"
	"github.com/ironsmile/nedomi/mock"
	"github.com/ironsmile/nedomi/storage"
	"github.com/ironsmile/nedomi/types"
	"github.com/ironsmile/nedomi/utils/testutils"
	"golang.org/x/net/context"
)

func min(l, r int) int {
	if l > r {
		return r
	}
	return l
}

type testApp struct {
	testing.TB
	up           *mock.RequestHandler
	cacheHandler *CachingProxy
	ctx          context.Context
	fsmap        map[string]string
	cleanup      func()
}

type fileInfo struct {
	path string
	size int
}

func (t *testApp) getFileName() string {
	return t.getFileSizes()[0].path // lazy
}

func (t *testApp) getFileSizes() []fileInfo {
	var files = make([]fileInfo, 0, len(t.fsmap))
	for key, value := range t.fsmap {
		files = append(files, fileInfo{
			path: key,
			size: len(value),
		})
	}
	return files
}

func (t *testApp) testRequest(req *http.Request, expected string, code int) {
	var rec = httptest.NewRecorder()
	t.cacheHandler.ServeHTTP(t.ctx, rec, req)
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

func (t *testApp) testRange(path string, begin, length uint64) {
	expected := t.fsmap[path]
	req := reqForRange(path, begin, length)
	t.testRequest(req, expected[begin:min(int(begin+length), len(expected))], http.StatusPartialContent)
}

func (t *testApp) testFullRequest(path string) {
	expected := t.fsmap[path]
	req, err := http.NewRequest("GET", "http://example.com/"+path, nil)
	if err != nil {
		panic(err)
	}

	t.testRequest(req, expected, http.StatusOK)
}

func newTestAppFromMap(t testing.TB, fsmap map[string]string) *testApp {
	up := mock.NewRequestHandler(fsMapHandler(fsmap))
	cpus := runtime.NumCPU()
	runtime.GOMAXPROCS(cpus)
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
		StorageObjects: 200,
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
		Scheduler: storage.NewScheduler(loc.Logger),
		Storage:   st,
	}

	cacheHandler, err := New(nil, loc, up)
	if err != nil {
		t.Fatal(err)
	}
	app := &testApp{
		TB:           t,
		up:           up,
		ctx:          context.Background(),
		cacheHandler: cacheHandler,
		fsmap:        fsmap,
		cleanup:      cleanup,
	}
	return app
}

func newTestApp(t testing.TB) *testApp {
	return newTestAppFromMap(t, generateFiles(10))
}
