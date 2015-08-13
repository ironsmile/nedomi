package disk

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/ironsmile/nedomi/types"
	"github.com/ironsmile/nedomi/upstream"
)

// Mock cache algorithm

type CacheAlgorithmMock struct{}

func (c *CacheAlgorithmMock) Lookup(o types.ObjectIndex) bool {
	return false
}

func (c *CacheAlgorithmMock) ShouldKeep(o types.ObjectIndex) bool {
	return false
}

func (c *CacheAlgorithmMock) AddObject(o types.ObjectIndex) error {
	return nil
}

func (c *CacheAlgorithmMock) PromoteObject(o types.ObjectIndex) {}

func (c *CacheAlgorithmMock) ConsumedSize() types.BytesSize {
	return 0
}

func (c *CacheAlgorithmMock) ReplaceRemoveChannel(ch chan<- types.ObjectIndex) {

}

func (c *CacheAlgorithmMock) Stats() types.CacheStats {
	return nil
}

// Mock http handler

type testHandler struct{}

func (t *testHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	for i := 0; i < 5; i++ {
		w.Header().Add(fmt.Sprintf("X-Header-%d", i), fmt.Sprintf("value-%d", i))
	}

	w.WriteHeader(200)
}

type fakeUpstream struct {
	upstream.Upstream
	responses map[string]fakeResponse
}

func (f *fakeUpstream) addFakeResponse(path string, fake fakeResponse) {
	f.responses[path] = fake
}

type fakeResponse struct {
	Status       string
	ResponseTime time.Duration
	Response     string
	Headers      http.Header
	err          error
}

func newFakeUpstream() *fakeUpstream {
	return &fakeUpstream{
		responses: make(map[string]fakeResponse),
	}
}

func (f *fakeUpstream) GetSize(path string) (int64, error) {
	fake, ok := f.responses[path]
	if !ok {
		return 0, nil // @todo fix
	}
	if fake.err != nil {
		return 0, fake.err
	}

	return int64(len(fake.Response)), nil
}

func (f *fakeUpstream) GetRequest(path string) (*http.Response, error) {
	end, _ := f.GetSize(path)
	return f.GetRequestPartial(path, 0, uint64(end))
}

func (f *fakeUpstream) GetRequestPartial(path string, start, end uint64) (*http.Response, error) {
	fake, ok := f.responses[path]
	if !ok {
		return nil, nil // @todo fix
	}
	if fake.ResponseTime > 0 {
		time.Sleep(fake.ResponseTime)
	}

	if fake.err != nil {
		return nil, fake.err
	}

	if length := uint64(len(fake.Response)); end > length {
		end = length
	}

	return &http.Response{
		Status: fake.Status,
		Body:   ioutil.NopCloser(bytes.NewBufferString(fake.Response[start:end])),
		Header: fake.Headers,
	}, nil
}

func (f *fakeUpstream) GetHeader(path string) (http.Header, error) {
	fake, ok := f.responses[path]
	if !ok {
		return nil, nil // @todo fix
	}
	if fake.ResponseTime > 0 {
		time.Sleep(fake.ResponseTime)
	}

	if fake.err != nil {
		return nil, fake.err
	}
	return fake.Headers, nil

}

type countingUpstream struct {
	*fakeUpstream
	called int32
}

func (c *countingUpstream) GetRequestPartial(path string, start, end uint64) (*http.Response, error) {
	atomic.AddInt32(&c.called, 1)
	return c.fakeUpstream.GetRequestPartial(path, start, end)
}
