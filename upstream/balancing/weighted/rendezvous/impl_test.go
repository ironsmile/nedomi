package rendezvous

import (
	"fmt"
	"math"
	"math/rand"
	"net/url"
	"testing"
	"time"

	"github.com/ironsmile/nedomi/types"
)

func TestPercentageCalculations(t *testing.T) {
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	numUpstreams := rnd.Intn(200)
	upstreams := make([]*types.UpstreamAddress, numUpstreams)

	for i := 0; i < numUpstreams; i++ {
		upstreams[i] = &types.UpstreamAddress{
			URL:    &url.URL{Host: fmt.Sprintf("upstream%d.com", i), Scheme: "http"},
			Weight: rnd.Uint32(),
		}
	}

	r := New()
	r.Set(upstreams)

	var totalPercent float64
	for _, b := range r.buckets {
		totalPercent += b.weightPercent
	}

	if math.Abs(totalPercent-1.0) > math.Nextafter(1.0, 2.0)-1.0 {
		t.Errorf("Bucket percentages do not combine to 1: %f", totalPercent)
	}
}
