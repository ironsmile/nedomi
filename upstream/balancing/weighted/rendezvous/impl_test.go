package rendezvous

import (
	"fmt"
	"math"
	"math/rand"
	"reflect"
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

func TestOperations(t *testing.T) {
	t.Parallel()

	upstreams := testutils.GetRandomUpstreams(50, 200)
	r := New()
	r.Set(upstreams)

	urlsToTest := rand.Intn(500)
	mapping := map[string]string{}
	for i := 0; i < urlsToTest; i++ {
		url := testutils.GenerateMeAString(rand.Int63(), 1+rand.Int63n(50))
		if res1, err := r.Get(url); err != nil {
			t.Errorf("Unexpected error when getting url %s: %s", url, err)
		} else if res2, err := r.Get(url); err != nil {
			t.Errorf("Unexpected error when getting url %s for the second time: %s", url, err)
		} else if !reflect.DeepEqual(res1, res2) {
			t.Errorf("The two results for url %s are different: %#v and %#v", url, res1, res2)
		} else {
			mapping[url] = res1.Host
		}
	}

	blackSheep := rand.Intn(len(upstreams))
	blackSheepsHost := upstreams[blackSheep].Host
	upstreams = upstreams[:blackSheep+copy(upstreams[blackSheep:], upstreams[blackSheep+1:])]
	r.Set(upstreams)
	for url, oldHost := range mapping {
		if res, err := r.Get(url); err != nil {
			t.Errorf("Unexpected error when getting url %s: %s", url, err)
		} else if oldHost != blackSheepsHost && res.Host != oldHost {
			t.Errorf("Expected to return the same old value %s for url %s but it returned %s",
				oldHost, url, res.Host)
			fmt.Printf("The black sheep was upstream #%d with host %s\n", blackSheep, blackSheepsHost)
		} else if oldHost == blackSheepsHost && res.Host == oldHost {
			t.Errorf("Expected to different value than the black sheep for url %s but it returned %s",
				url, res.Host)
		}
	}
}
