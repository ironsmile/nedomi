package balancing

import (
	"testing"

	"github.com/ironsmile/nedomi/utils/testutils"
)

const (
	upstremsCount = 5
	urlLength     = 50
)

func runTest(b *testing.B, id string) {
	b.ReportAllocs()

	inst := allAlgorithms[id]()
	inst.Set(testutils.GetUpstreams(1, upstremsCount))

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			inst.Get(testutils.GenerateMeAString(0, urlLength))
		}
	})
}

func BenchmarkKetama(b *testing.B)               { runTest(b, "ketama") }
func BenchmarkLegacyKetama(b *testing.B)         { runTest(b, "legacyketama") }
func BenchmarkRandom(b *testing.B)               { runTest(b, "random") }
func BenchmarkRendezvous(b *testing.B)           { runTest(b, "rendezvous") }
func BenchmarkUnweightedRandom(b *testing.B)     { runTest(b, "unweighted-random") }
func BenchmarkUnweightedRoundRobin(b *testing.B) { runTest(b, "unweighted-roundrobin") }
