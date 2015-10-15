package types

// UpstreamBalancingAlgorithm encapulates thread-safe methods that are used for
// balancing user requests between a set of upstream addresses. That is done
// according to the specific balancing algorithm implementation.
type UpstreamBalancingAlgorithm interface {

	// Set is used to specify all the possible upstream addresses the algorithm
	// has to choose from.
	Set([]*UpstreamAddress)

	// Get returns a specific address, according to the supplied path.
	Get(string) (*UpstreamAddress, error)
}
