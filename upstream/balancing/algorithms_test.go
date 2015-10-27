package balancing

import (
	"fmt"
	"math/rand"
	"net/url"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/ironsmile/nedomi/types"
	"github.com/ironsmile/nedomi/utils/testutils"
)

func init() {
	seed := time.Now().UnixNano()
	fmt.Printf("Initializing with random seed %d\n", seed)
	rand.Seed(seed)
}

func getUpstream(i int) *types.UpstreamAddress {
	return &types.UpstreamAddress{
		URL:         url.URL{Host: fmt.Sprintf("127.0.%d.%d", (i/256)%256, i%256), Scheme: "http"},
		Hostname:    fmt.Sprintf("www.upstream%d.com", i),
		Port:        "80",
		OriginalURL: &url.URL{Host: fmt.Sprintf("www.upstream%d.com", i), Scheme: "http"},
		Weight:      1 + uint32(rand.Intn(1000)),
	}
}

func getRandomUpstreams(minCount, maxCount int) []*types.UpstreamAddress {
	count := minCount + rand.Intn(maxCount-minCount+1)
	result := make([]*types.UpstreamAddress, count)
	for i := 0; i < count; i++ {
		result[i] = getUpstream(i)
	}
	return result
}

func TestBasicOperations(t *testing.T) {
	t.Parallel()

	for id, algo := range allAlgorithms {
		inst := algo()
		if res, err := inst.Get("bogus1"); err == nil {
			t.Errorf("Expected to receive error for unitialized algorithm %s but received %#v", id, res)
		}
		upstream := getUpstream(rand.Int())
		inst.Set([]*types.UpstreamAddress{upstream})
		if res, err := inst.Get("bogus2"); err != nil {
			t.Errorf("Received an unexpected error for algorithm %s: %s", id, err)
		} else if !reflect.DeepEqual(res, upstream) {
			t.Errorf("Algorithm %s returned '%#v' when it should have returned '%#v'", id, res, upstream)
		}
		inst.Set([]*types.UpstreamAddress{})
		if res, err := inst.Get("bogus3"); err == nil {
			t.Errorf("Expected to receive error for empty algorithm %s but received %#v", id, res)
		}
	}
}

// This test will probably only be useful if `go test -race` is used
func TestRandomConcurrentUsage(t *testing.T) {
	t.Parallel()
	wg := sync.WaitGroup{}

	randomlyTestAlgorithm := func(id string, inst types.UpstreamBalancingAlgorithm) {
		inst.Set([]*types.UpstreamAddress{getUpstream(rand.Int())}) // Prevent expected errors

		setters := 50 + rand.Intn(200)
		wg.Add(setters)
		for i := 0; i < setters; i++ {
			go func() {
				time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)
				inst.Set(getRandomUpstreams(1, 100))
				wg.Done()
			}()
		}

		getters := 100 + rand.Intn(500)
		wg.Add(getters)
		for i := 0; i < getters; i++ {
			go func() {
				time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)
				path := fmt.Sprintf("some/path/%d.jpn", rand.Int())
				if _, err := inst.Get(path); err != nil {
					t.Errorf("Unexpected algorithm %s error for path %s: %s", id, path, err)
				}
				wg.Done()
			}()
		}
	}

	for id, constructor := range allAlgorithms {
		randomlyTestAlgorithm(id, constructor())
	}

	wg.Wait()
}

func TestConsistentHashAlgorithms(t *testing.T) {
	//t.Parallel()
	wg := sync.WaitGroup{}

	testConsistenHashAlgorithm := func(id string, inst types.UpstreamBalancingAlgorithm) {
		upstreams := getRandomUpstreams(50, 200)
		inst.Set(upstreams)

		urlsToTest := rand.Intn(500)
		mapping := map[string]string{}
		for i := 0; i < urlsToTest; i++ {
			url := testutils.GenerateMeAString(rand.Int63(), 1+rand.Int63n(50))
			if res1, err := inst.Get(url); err != nil {
				t.Errorf("Unexpected error when getting url %s from algorithm %s: %s", url, id, err)
			} else if res2, err := inst.Get(url); err != nil {
				t.Errorf("Unexpected error when getting url %s from algorithm %s for the second time: %s", url, id, err)
			} else if !reflect.DeepEqual(res1, res2) {
				t.Errorf("The two results for url %s by algorithm %s are different: %#v and %#v", url, id, res1, res2)
			} else {
				mapping[url] = res1.Host
			}
		}

		//!TODO: enable this check again
		/*newUpstream := getUpstream(len(upstreams))
		upstreams = append(upstreams, newUpstream)
		inst.Set(upstreams)
		for url, oldHost := range mapping {
			if res, err := inst.Get(url); err != nil {
				t.Errorf("Unexpected error when getting url %s from algorithm %s: %s", url, id, err)
			} else if res.Host != newUpstream.Host && res.Host != oldHost {
				t.Errorf("Expected algorithm %s to return either %s or %s for url %s but it returned %s",
					id, newUpstream.Host, oldHost, url, res.Host)
			} else {
				mapping[url] = res.Host
			}
		}*/

		blackSheep := rand.Intn(len(upstreams))
		blackSheepsHost := upstreams[blackSheep].Host
		inst.Set(append(upstreams[:blackSheep], upstreams[blackSheep+1:]...))
		for url, oldHost := range mapping {
			if res, err := inst.Get(url); err != nil {
				t.Errorf("Unexpected error when getting url %s from algorithm %s: %s", url, id, err)
			} else if oldHost != blackSheepsHost && res.Host != oldHost {
				t.Errorf("Expected algorithm %s to return the same old value %s for url %s but it returned %s",
					id, oldHost, url, res.Host)
			} else if oldHost == blackSheepsHost && res.Host == oldHost {
				t.Errorf("Expected algorithm %s to different value than the black sheep for url %s but it returned %s",
					id, url, res.Host)
			}
		}

		wg.Done()
	}

	algorithmsToTest := []string{"ketama", "rendezvous"} //!TODO: add "unweighted-jump"
	for _, id := range algorithmsToTest {
		wg.Add(1)
		testConsistenHashAlgorithm(id, allAlgorithms[id]())
	}

	wg.Wait()
}
