package balancing

import (
	"fmt"
	"math"
	"math/rand"
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

func TestBasicOperations(t *testing.T) {
	t.Parallel()

	for id, algo := range allAlgorithms {
		inst := algo()
		if res, err := inst.Get("bogus1"); err == nil {
			t.Errorf("Expected to receive error for unitialized algorithm %s but received %#v", id, res)
		}
		upstream := testutils.GetUpstream(rand.Int())
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
		inst.Set([]*types.UpstreamAddress{testutils.GetUpstream(rand.Int())}) // Prevent expected errors

		setters := 50 + rand.Intn(200)
		wg.Add(setters)
		for i := 0; i < setters; i++ {
			go func() {
				time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)
				inst.Set(testutils.GetRandomUpstreams(1, 100))
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

func getTotalWeight(upstreams []*types.UpstreamAddress) float64 {
	var res float64
	for _, u := range upstreams {
		res += float64(u.Weight)
	}
	return res
}

func TestConsistentHashAlgorithms(t *testing.T) {
	t.Skip() //!TODO: enable test
	t.Parallel()
	wg := sync.WaitGroup{}

	testConsistenHashAlgorithm := func(id string, inst types.UpstreamBalancingAlgorithm) {
		upstreams := testutils.GetRandomUpstreams(3, 100)
		inst.Set(upstreams)

		urlsToTest := 15000 + rand.Intn(5000)
		mapping := map[string]string{}
		for i := 0; i < urlsToTest; i++ {
			url := testutils.GenerateMeAString(rand.Int63(), 5+rand.Int63n(100))
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

		checkDeviation := func(newUpstreams []*types.UpstreamAddress) {
			inst.Set(newUpstreams)
			var sameCount float64
			for url, oldHost := range mapping {
				if res, err := inst.Get(url); err != nil {
					t.Errorf("Unexpected error when getting url %s from algorithm %s: %s", url, id, err)
				} else if res.Host == oldHost {
					sameCount++
				}
			}
			total := float64(len(mapping))
			oldWeght := getTotalWeight(upstreams)
			newWeght := getTotalWeight(newUpstreams)
			weightDiff := math.Abs(oldWeght - newWeght)

			//!TODO: do not use a fixed threshold
			if math.Abs(weightDiff/oldWeght-(total-sameCount)/total) > 0.10 {
				t.Errorf("[%s] Same count is %f of %f (count diff %f%%); upstreams are %d out of %d (weight diff %f-%f=%f or %f%%); deviation from expected: %f\n\n",
					id, sameCount, total, (total-sameCount)*100/total,
					len(newUpstreams), len(upstreams), oldWeght, newWeght, oldWeght-newWeght, weightDiff*100/oldWeght,
					math.Abs(weightDiff/oldWeght-(total-sameCount)/total)*100)
			}
		}

		// Add an extra server at the start
		//!TODO
		// Remove a random server
		randomServer := rand.Intn(len(upstreams))
		checkDeviation(append(upstreams[:randomServer], upstreams[randomServer+1:]...))
		// Add an extra server at that point
		//!TODO
		// Add an extra server at the end
		checkDeviation(append(upstreams, testutils.GetUpstream(len(upstreams))))
		wg.Done()
	}

	algorithmsToTest := []string{"ketama", "rendezvous"}
	for _, id := range algorithmsToTest {
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go testConsistenHashAlgorithm(id, allAlgorithms[id]())
		}
	}

	wg.Wait()
}
