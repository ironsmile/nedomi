package cache

import (
	"math/rand"
	"net/http"
	"net/http/httptest"
	"testing"

	"golang.org/x/net/context"

	"github.com/ironsmile/nedomi/types"
	"github.com/ironsmile/nedomi/utils"
	"github.com/ironsmile/nedomi/utils/testutils"
)

func BenchmarkStorageSimultaneousRangeGetsFillingUp(b *testing.B) {
	app := newTestApp(b)
	defer app.cleanup()
	var files = app.getFileSizes()
	var filesCount = len(files)
	testfunc := func(index int) {
		file := files[(index)%filesCount]
		var begin = rand.Intn(file.size - 4)
		var length = rand.Intn(file.size-begin-1) + 2
		app.testRange(file.path, uint64(begin), uint64(length))
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for i := 0; pb.Next(); i++ {
			testfunc(i)
		}
	})
}

func benchmarkGetContentsiOf40(b *testing.B, i int) {
	app := newTestApp(b)
	defer app.cleanup()
	var file = "40parts"
	var partSize = app.cacheHandler.Cache.Storage.PartSize()
	app.fsmap[file] = testutils.GenerateMeAString(1, int64(40*partSize))

	app.testRange(file, 0, partSize)
	req, err := http.NewRequest("GET", "http://example.com/"+file, nil)
	if err != nil {
		panic(err)
	}
	var resp = httptest.NewRecorder()

	objID := app.cacheHandler.NewObjectIDForURL(req.URL)
	indexes := utils.BreakInIndexes(
		objID, 0,
		40*partSize-1,
		partSize)
	ctx := context.Background()

	testfunc := func() {
		rq := &reqHandler{
			CachingProxy: app.cacheHandler,
			ctx:          ctx,
			req:          req,
			resp:         resp,
			objID:        objID,
			reqID:        types.RequestID(`testiID`),
		}
		rq.obj, _ = rq.Cache.Storage.GetMetadata(rq.objID)
		contents, _, err := rq.getContents(indexes, i)
		if err != nil {
			b.Fatal(err)
		}
		if err := contents.Close(); err != nil {
			b.Fatal(err)
		}
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			testfunc()
		}
	})
}

func BenchmarkGetContents0Of40(b *testing.B) {
	benchmarkGetContentsiOf40(b, 0)
}

func BenchmarkGetContents1Of40(b *testing.B) {
	benchmarkGetContentsiOf40(b, 1)
}
