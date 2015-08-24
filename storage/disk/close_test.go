package disk

import (
	"io/ioutil"
	"net/http"
	"runtime"
	"testing"

	"github.com/ironsmile/nedomi/contexts"
	"github.com/ironsmile/nedomi/types"
	"golang.org/x/net/context"
)

func TestCloseOnNotUsedDiskWorks(t *testing.T) {
	_, cz, ca, _ := setup()
	storage := New(cz, ca, newStdLogger())
	err := storage.Close()
	if err != nil {
		t.Errorf("Storage.Close() shouldn't have returned error\n%s", err)
	}
}

func TestDiskCloseReturnesAfterFinishingRequests(t *testing.T) {
	expected := "awesome"
	fakeup, cz, ca, _ := setup()
	runtime.GOMAXPROCS(runtime.NumCPU())
	up := newBlockingUpstream(fakeup)
	up.addFakeResponse("/path",
		fakeResponse{
			Status:       "200",
			ResponseTime: 0,
			Response:     expected,
		})

	storage := New(cz, ca, newStdLogger())
	ctx := contexts.NewVhostContext(context.Background(), &types.VirtualHost{Upstream: up})
	wCh := make(chan struct{})
	go func(ch chan struct{}) {
		defer close(ch)
		oid := types.ObjectID{}
		oid.CacheKey = "1"
		oid.Path = "/path"
		resp, err := storage.GetFullFile(ctx, oid)
		if err != nil {
			t.Errorf("Got Error on a GetFullFile on closing storage:\n%s", err)
		}
		b, err := ioutil.ReadAll(resp)
		if err != nil {
			t.Errorf("Got Error on a ReadAll on an already closed storage:\n%s", err)
		}
		if string(b) != expected {
			t.Errorf("Expected read from GetFullFile was %s but got %s", expected, string(b))
		}

	}(wCh)
	go close(<-up.requestPartial)
	err := storage.Close()
	if err != nil {
		t.Errorf("Storage.Close() shouldn't have returned error\n%s", err)
	}
	<-wCh
}

func TestDiskCloseReturnesAfterFinishingHeaders(t *testing.T) {
	var expected = make(http.Header)
	var expectedHeader, expectedValue = "awesome-header", "awesome-value"
	expected.Add(expectedHeader, expectedValue)
	fakeup, cz, ca, _ := setup()
	runtime.GOMAXPROCS(runtime.NumCPU())
	up := newBlockingUpstream(fakeup)
	up.addFakeResponse("/path",
		fakeResponse{
			Status:       "200",
			ResponseTime: 0,
			Headers:      expected,
		})

	storage := New(cz, ca, newStdLogger())
	ctx := contexts.NewVhostContext(context.Background(), &types.VirtualHost{Upstream: up})
	wCh := make(chan struct{})
	go func(ch chan struct{}) {
		defer close(ch)
		oid := types.ObjectID{}
		oid.CacheKey = "1"
		oid.Path = "/path"

		headers, err := storage.Headers(ctx, oid)
		if err != nil {
			t.Errorf("Got Error on a Headers on closing storage:\n%s", err)
		}
		if headers.Get(expectedHeader) != expectedValue { //!TODO: better check
			t.Errorf("The returned headers:\n%s\n Did not match the expected headers\n%s", headers, expected)
		}

	}(wCh)
	go close(<-up.requestHeader)
	err := storage.Close()
	if err != nil {
		t.Errorf("Storage.Close() shouldn't have returned error\n%s", err)
	}
	<-wCh
}
