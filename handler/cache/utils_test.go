package cache

import (
	"io"
	"math/rand"
	"net/http"
	"testing"
	"time"

	"github.com/ironsmile/nedomi/config"
	"github.com/ironsmile/nedomi/logger"
	"github.com/ironsmile/nedomi/types"
	"github.com/ironsmile/nedomi/utils/httputils"
	"github.com/ironsmile/nedomi/utils/testutils"

	"golang.org/x/tools/godoc/vfs/httpfs"
	"golang.org/x/tools/godoc/vfs/mapfs"
)

func init() {
	rand.Seed(time.Now().Unix())
}

func fsMapHandler(fsmap map[string]string) http.HandlerFunc {
	return func(wr http.ResponseWriter, req *http.Request) {
		wr.Header().Add("Expires", time.Now().Add(time.Hour).Format(time.RFC1123))
		http.FileServer(httpfs.New(mapfs.New(fsmap))).ServeHTTP(wr, req)
	}
}

func generateFiles(n int) map[string]string {
	var files = make(map[string]string, n)
	for i := 0; n > i; i++ {
		name := testutils.GenerateMeAString(int64(i), 5)
		files[name] = testutils.GenerateMeAString(int64(i*n), rand.Int63n(500)+200)
	}
	return files
}

func newStdLogger() types.Logger {
	l, err := logger.New(config.NewLogger("std", []byte(`{"level":"error"}`)))
	if err != nil {
		panic(err)
	}
	return l
}

func read(t *testing.T, r io.Reader, b []byte) {
	n, err := r.Read(b)
	if err != nil {
		t.Errorf("error reading %d bytes from %+v - %s",
			len(b), r, err)
	}
	if n != len(b) {
		t.Errorf("read %d bytes from %+v  instead of %d",
			n, r, len(b))
	}
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
