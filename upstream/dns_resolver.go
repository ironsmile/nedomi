package upstream

import (
	"github.com/ironsmile/nedomi/config"
	"github.com/ironsmile/nedomi/types"
	"github.com/ironsmile/nedomi/upstream/balancing"
)

func initDNSResolver(algo balancing.Algorithm, addresses []config.UpstreamAddress) {
	//!TODO: pass cancel channel
	//!TODO: implement an intelligent TTL-aware persistent resolver
	resolved := []types.UpstreamAddress{}

	for _, addr := range addresses {
		//!TODO: actually resolve :)
		resolved = append(resolved, types.UpstreamAddress{URL: addr.URL})
	}

	fillMissingWeights(resolved)

	algo.Set(resolved)
}
