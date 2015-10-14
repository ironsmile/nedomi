package balancing

import "github.com/ironsmile/nedomi/types"

// Algorithm encapulates thread-safe methods that are used for balancing user
// requests between the set of upstream addresses according to the specific
// algorithm .
type Algorithm interface {
	Set([]*types.UpstreamAddress)

	Get(string) (*types.UpstreamAddress, error)
}
