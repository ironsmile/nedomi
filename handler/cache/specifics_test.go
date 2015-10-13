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

type testApp struct {
	testing.TB
	up           *mock.RequestHandler
	cacheHandler *CachingProxy
	ctx          context.Context
	fsmap        map[string]string
	cleanup      func()
}

func (a *testApp) getFileName() string {
	return a.getFileSizes()[0].path // lazy
}

type fileInfo struct {
	path string
	size int
}

func (a *testApp) getFileSizes() []fileInfo {
	var files = make([]fileInfo, 0, len(a.fsmap))
	for key, value := range a.fsmap {
		files = append(files, fileInfo{
			path: key,
			size: len(value),
		})
	}
	return files
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
	expected := t.fsmap[path]
	req := reqForRange(path, begin, length)
	testRequest(t, req, expected[begin:begin+length], http.StatusPartialContent)
}

func testFullRequest(t *testApp, path string) {
	expected := t.fsmap[path]
	req, err := http.NewRequest("GET", "http://example.com/"+path, nil)
	if err != nil {
		panic(err)
	}

	testRequest(t, req, expected, http.StatusOK)
}

func TestVeryFragmentedFile(t *testing.T) {
	t.Parallel()
	app := realerSetup(t)
	var file = "long"
	app.fsmap[file] = generateMeAString(1, 2000)
	defer app.cleanup()

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

func realerSetupFromMap(t testing.TB, fsmap map[string]string) *testApp {
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
		Scheduler: storage.NewScheduler(),
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

func realerSetup(t testing.TB) *testApp {
	return realerSetupFromMap(t, generateFiles(10))
}

func Test2PartsFile(t *testing.T) {
	var fsmap = make(map[string]string)
	var file = "2parts"
	fsmap[file] = generateMeAString(2, 10)
	t.Parallel()
	app := realerSetupFromMap(t, fsmap)
	defer app.cleanup()
	testRange(app, file, 2, 8)
	testFullRequest(app, file)
}
