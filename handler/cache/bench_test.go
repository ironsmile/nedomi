package cache

import (
	"math/rand"
	"runtime"
	"testing"
)

func BenchmarkStorageSimultaneousRangeGetsFillingUp(b *testing.B) {
	var filesCount = runtime.NumCPU() * 10

	app := newTestApp(b)
	defer app.cleanup()
	var files = app.getFileSizes()
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
