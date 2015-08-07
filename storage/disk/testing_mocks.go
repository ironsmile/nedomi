package disk

import (
	"fmt"
	"net/http"

	"github.com/ironsmile/nedomi/config"
	"github.com/ironsmile/nedomi/types"
)

// Mock cache manager

type cacheManagerMock struct{}

func (c *cacheManagerMock) Lookup(o types.ObjectIndex) bool {
	return false
}

func (c *cacheManagerMock) ShouldKeep(o types.ObjectIndex) bool {
	return false
}

func (c *cacheManagerMock) AddObject(o types.ObjectIndex) error {
	return nil
}

func (c *cacheManagerMock) PromoteObject(o types.ObjectIndex) {}

func (c *cacheManagerMock) ConsumedSize() config.BytesSize {
	return 0
}

func (c *cacheManagerMock) ReplaceRemoveChannel(ch chan<- types.ObjectIndex) {

}

func (c *cacheManagerMock) Stats() types.CacheStats {
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
