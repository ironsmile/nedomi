package cache

import (
	"math/rand"
	"runtime"
	"testing"

	"golang.org/x/net/context"
)

func generateFiles(n int) []string {
	var names = make([]string, n)
	for i := 0; n > i; i++ {
		names[i] = generateMeAString(int64(i), 5)
		fsmap[names[i]] = generateMeAString(int64(i*n), int64(i*100)%5000+200)
	}
	return names
}

func BenchmarkStorageSimultaneousRangeGetsFillingUp(b *testing.B) {
	var filesCount = runtime.NumCPU() * 10
	var files = generateFiles(filesCount)

	_, loc, _, _, cleanup := realerSetup(b)
	defer cleanup()
	cacheHandler, err := New(nil, loc, nil)
	if err != nil {
		b.Fatal(err)
	}
	app := &testApp{
		TB:           b,
		ctx:          context.Background(),
		cacheHandler: cacheHandler,
	}

	testfunc := func(index int) {
		file := files[(index)%filesCount]
		expected := fsmap[file]
		var begin = rand.Intn(len(expected) - 4)
		var length = rand.Intn(len(expected)-begin-1) + 2
		testRange(app, file, uint64(begin), uint64(length))
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for i := 0; pb.Next(); i++ {
			testfunc(i)
		}
	})
}
