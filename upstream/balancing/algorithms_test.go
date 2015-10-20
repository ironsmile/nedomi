package balancing

import "testing"

func TestExpectedErrors(t *testing.T) {
	//!TODO: test all possible balancing algorithms if they return errors with
	// no configured upstreams and if they do not return errors with at least
	// one upstream
}

func TestRandomConcurrentUsage(t *testing.T) {
	//!TODO: test all possible balancing algorithms if they have some issues
	// being used concurrently. We won't check results, just see if something
	// panics or the race condition detector complains
}

func TestConsistentHashAlgorithms(t *testing.T) {
	//!TODO: test all balancing algorithms that are based on consistent hashing
	// (unweighted-jump, ketama, rendezvous) - we will generate random upstreams
	// and urls and verify that:
	// 1. If we do not change the upstreams, they always return the same result
	// for a specific key
	// 2. If we add a new upstream, the algorithm returns either the old result
	// or the newly added upstream for all keys.
	// 3. If we remove an upstream:
	//   - All paths that returned other upstreams before still return them
	//   - All paths that returned the deleted upstream before now return another one
	// Also, we may want to verify that the order of the upstram slice does not
	// matter? If so, some changes to the implementations will be required.
}
