package throttle

import "testing"

func BenchmarkThrottledWriterWithReadFrom(b *testing.B) {
	benchInParallel(b, testResponseWriterWithReadFrom)
}

func BenchmarkThrottledWriter(b *testing.B) {
	benchInParallel(b, testResponseWriter)
}

func benchInParallel(b *testing.B, f testFunc) {
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			runInParallel(b, benchTests[:], f)
		}
	})
}

var benchTestsPre = []throttleTest{
	{"20M with 5MB/s", content["20M"], 5 * 1024 * 1024}, // 4 seconds
	// {content["2M"], 200 * 1024},       // 10 seconds
	{"2M with 1MB/s", content["2M"], 1024 * 1024},       // 2 seconds
	{"10K with 5MB/s", content["10K"], 5 * 1024 * 1024}, // 2 seconds
	//{content["10K"], 1024 * 1024},     // 10 seconds
	//{content["1K"], 100},              // 10+ seconds
}

var benchTests = func(tests []throttleTest, repeat int) []throttleTest {
	var result = make([]throttleTest, repeat*len(tests))
	for i := 0; repeat > i; i++ {
		copy(result[i*len(tests):], tests)
	}

	return result
}(benchTestsPre, 10) // LOAD
