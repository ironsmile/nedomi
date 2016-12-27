package cache

import (
	"io"
	"net/http"
	"syscall"
	"testing"

	"github.com/ironsmile/nedomi/utils/httputils"
)

func TestTooManyFiles(t *testing.T) {
	// because of the usage of Rlimit - parallel-ing this test
	// could lead to the failure of others

	app := newTestApp(t)
	defer app.cleanup()
	var nofileRlimit = new(syscall.Rlimit)
	if err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, nofileRlimit); err != nil {
		t.Fatal("Errof on syscall.GetRlimit", err)
	}
	defer func(oldlimit syscall.Rlimit) {
		if err := syscall.Setrlimit(syscall.RLIMIT_NOFILE, &oldlimit); err != nil {
			t.Fatal("Errof on syscall.SetRlimit", err)
		}
	}(*nofileRlimit)
	var file = app.getFileName()

	app.testFullRequest(file) // this should succeed

	nofileRlimit.Cur = 0
	if err := syscall.Setrlimit(syscall.RLIMIT_NOFILE, nofileRlimit); err != nil {
		t.Fatal("Errof on syscall.SetRlimit", err)
	}
	req, err := http.NewRequest("GET", "http://example.com/"+file, nil)
	if err != nil {
		t.Fatal(err)
	}
	app.testRequest(req, http.StatusText(http.StatusInternalServerError)+"\n", http.StatusInternalServerError)
}

func TestTooManyFilesInTheMiddle(t *testing.T) {
	// because of the usage of Rlimit - parallel-ing this test
	// could lead to the failure of others

	app := newTestApp(t)
	defer app.cleanup()
	var nofileRlimit = new(syscall.Rlimit)
	if err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, nofileRlimit); err != nil {
		t.Fatal("Errof on syscall.GetRlimit", err)
	}
	defer func(oldlimit syscall.Rlimit) {
		if err := syscall.Setrlimit(syscall.RLIMIT_NOFILE, &oldlimit); err != nil {
			t.Fatal("Errof on syscall.SetRlimit", err)
		}
	}(*nofileRlimit)
	var file = app.getFileName()

	app.testFullRequest(file) // this should succeed

	req, err := http.NewRequest("GET", "http://example.com/"+file, nil)
	if err != nil {
		t.Fatal(err)
	}

	r, w := io.Pipe()
	rec := httputils.NewFlexibleResponseWriter(func(frw *httputils.FlexibleResponseWriter) {
		frw.BodyWriter = w
	})
	go app.cacheHandler.ServeHTTP(app.ctx, rec, req)
	var buf [32]byte
	read(t, r, buf[:2]) // read some
	// make it fail on next part
	nofileRlimit.Cur = 0
	if err := syscall.Setrlimit(syscall.RLIMIT_NOFILE, nofileRlimit); err != nil {
		t.Fatal("Errof on syscall.SetRlimit", err)
	}

	read(t, r, buf[:2]) // read some more
	n, _ := r.Read(buf[:32])
	if n != 1 {
		t.Errorf("expected to read 1 not %d", n)
	}
}
