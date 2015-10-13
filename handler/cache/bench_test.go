package cache

import (
	"math/rand"
	"runtime"
	"testing"
)

func generateFiles(n int) map[string]string {
	var files = make(map[string]string, n)
	for i := 0; n > i; i++ {
		name := generateMeAString(int64(i), 5)
		files[name] = generateMeAString(int64(i*n), rand.Int63n(500)+200)
	}
	return files
}

func BenchmarkStorageSimultaneousRangeGetsFillingUp(b *testing.B) {
	var filesCount = runtime.NumCPU() * 10

	app := realerSetup(b)
	defer app.cleanup()
	var files = app.getFileSizes()
	testfunc := func(index int) {
		file := files[(index)%filesCount]
		var begin = rand.Intn(file.size - 4)
		var length = rand.Intn(file.size-begin-1) + 2
		testRange(app, file.path, uint64(begin), uint64(length))
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for i := 0; pb.Next(); i++ {
			testfunc(i)
		}
	})
}
