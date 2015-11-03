package rendezvous

import (
	"fmt"
	"math"
	"math/rand"
	"testing"
	"time"

	"github.com/ironsmile/nedomi/utils/testutils"
)

func init() {
	seed := time.Now().UnixNano()
	fmt.Printf("Initializing with random seed %d\n", seed)
	rand.Seed(seed)
}

func TestPercentageCalculations(t *testing.T) {
	t.Parallel()

	r := New()
	r.Set(testutils.GetRandomUpstreams(1, 200))

	var totalPercent float64
	for _, b := range r.buckets {
		totalPercent += b.weightPercent
	}

	if math.Abs(totalPercent-1.0) > 0.000001 {
		t.Errorf("Bucket percentages do not combine to 1: %f", totalPercent)
	}
}
