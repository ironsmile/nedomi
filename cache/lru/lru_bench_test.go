package lru

import (
	"math/rand"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ironsmile/nedomi/config"
	"github.com/ironsmile/nedomi/logger"
	"github.com/ironsmile/nedomi/types"
	"github.com/ironsmile/nedomi/utils/testutils"
)

const benchCacheSize = 1 << 20

var randUint32 = func() func() uint32 {
	r := rand.New(rand.NewSource(time.Now().Unix()))
	ch := make(chan uint32, 1000)
	go func() {
		for {
			ch <- r.Uint32()
		}
	}()
	return func() uint32 {
		return <-ch
	}
}()

func aFullCache(b *testing.B, size uint64) *TieredLRUCache {
	cz := getCacheZone()
	cz.StorageObjects = size
	l, _ := logger.New(config.NewLogger("nillogger", nil))
	lru := New(cz, mockRemove, l)
	fillCache(b, lru)
	return lru

}
func BenchmarkFilling(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			lru := aFullCache(b, benchCacheSize)
			if lru.stats().Objects() != benchCacheSize {
				b.Fatalf("expected %d objects in the full lru but got %d",
					benchCacheSize,
					lru.stats().Objects())
			}
		}
	})
}

var randPath = func() func() string {
	ch := make(chan string, 1000)
	go func() {
		for seed := int64(0); ; seed++ {
			ch <- testutils.GenerateMeAString(seed, 20)
		}
	}()
	return func() string {
		return <-ch
	}
}()

func BenchmarkLookupAndRemove(b *testing.B) {
	b.StopTimer()
	lru := aFullCache(b, benchCacheSize)
	var tooFast uint64
	b.StartTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			oi := getObjectIndexFor(randUint32(), "1.1", randPath())
			lru.AddObject(oi)
			if !lru.Lookup(oi) {
				atomic.AddUint64(&tooFast, 1)
			}
		}
	})
	b.Logf("%d times Lookup didn't lookup fast enough an object it just promoted", tooFast)
}

func getObjectIndexFor(part uint32, key, path string) *types.ObjectIndex {
	return &types.ObjectIndex{
		ObjID: types.NewObjectID(key, path),
		Part:  part,
	}
}

func BenchmarkResizeInHalf(b *testing.B)   { benchResize(b, benchCacheSize, benchCacheSize/2) }
func BenchmarkResizeInQuater(b *testing.B) { benchResize(b, benchCacheSize, benchCacheSize/4) }
func BenchmarkResizeByQuater(b *testing.B) { benchResize(b, benchCacheSize, (benchCacheSize/4)*3) }

func benchResize(b *testing.B, startingSize, endSize uint64) {
	for i := 0; b.N > i; i++ {
		b.StopTimer()
		lru := aFullCache(b, startingSize)
		b.StartTimer()
		lru.ChangeConfig(1, benchCacheSize, endSize)
	}
}
