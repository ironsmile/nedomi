package cache

import (
	"io"
	"net/http"
	"syscall"
	"testing"

	"github.com/ironsmile/nedomi/utils/httputils"

	"golang.org/x/net/context"
)

func TestCachingProxyHandler(t *testing.T) {
	t.Parallel()
	t.Skip("TODO: write tests for the handler and all of its helpers")
}

func TestTooManyFiles(t *testing.T) {
	// because of the usage of Rlimit - parallel-ing this test
	// could lead to the failure of others

	var file = "tooManyFiles"
	fsmap[file] = generateMeAString(1, 50)
	up, loc, _, _, cleanup := realerSetup(t)
	defer cleanup()
	var nofileRlimit = new(syscall.Rlimit)
	if err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, nofileRlimit); err != nil {
		t.Fatal("Errof on syscall.GetRlimit", err)
	}
	defer func(oldlimit syscall.Rlimit) {
		if err := syscall.Setrlimit(syscall.RLIMIT_NOFILE, &oldlimit); err != nil {
			t.Fatal("Errof on syscall.SetRlimit", err)
		}
	}(*nofileRlimit)
	cacheHandler, err := New(nil, loc, up)
	if err != nil {
		t.Fatal(err)
	}
	app := &testApp{
		TB:           t,
		ctx:          context.Background(),
		cacheHandler: cacheHandler,
	}
	testFullRequest(app, file) // this should succeed

	nofileRlimit.Cur = 0
	if err := syscall.Setrlimit(syscall.RLIMIT_NOFILE, nofileRlimit); err != nil {
		t.Fatal("Errof on syscall.SetRlimit", err)
	}
	req, err := http.NewRequest("GET", "http://example.com/"+file, nil)
	if err != nil {
		t.Fatal(err)
	}
	testRequest(app, req, http.StatusText(http.StatusInternalServerError)+"\n", http.StatusInternalServerError)
}

func TestTooManyFilesInTheMiddle(t *testing.T) {
	// because of the usage of Rlimit - parallel-ing this test
	// could lead to the failure of others

	var file = "tooManyFiles"
	fsmap[file] = generateMeAString(1, 50)
	up, loc, _, _, cleanup := realerSetup(t)
	defer cleanup()
	var nofileRlimit = new(syscall.Rlimit)
	if err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, nofileRlimit); err != nil {
		t.Fatal("Errof on syscall.GetRlimit", err)
	}
	defer func(oldlimit syscall.Rlimit) {
		if err := syscall.Setrlimit(syscall.RLIMIT_NOFILE, &oldlimit); err != nil {
			t.Fatal("Errof on syscall.SetRlimit", err)
		}
	}(*nofileRlimit)
	cacheHandler, err := New(nil, loc, up)
	if err != nil {
		t.Fatal(err)
	}
	app := &testApp{
		TB:           t,
		ctx:          context.Background(),
		cacheHandler: cacheHandler,
	}
	testFullRequest(app, file) // this should succeed

	req, err := http.NewRequest("GET", "http://example.com/"+file, nil)
	if err != nil {
		t.Fatal(err)
	}

	r, w := io.Pipe()
	rec := httputils.NewFlexibleResponseWriter(func(frw *httputils.FlexibleResponseWriter) {
		frw.BodyWriter = w
	})
	go app.cacheHandler.RequestHandle(app.ctx, rec, req)
	var buf [20]byte
	read(t, r, buf[:2]) // read some
	// make it fail on next part
	nofileRlimit.Cur = 0
	if err := syscall.Setrlimit(syscall.RLIMIT_NOFILE, nofileRlimit); err != nil {
		t.Fatal("Errof on syscall.SetRlimit", err)
	}

	read(t, r, buf[:2]) // read some more
	n, err := r.Read(buf[:2])
	if n != 1 {
		t.Errorf("expected to read 1 not %d", n)
	}
}
