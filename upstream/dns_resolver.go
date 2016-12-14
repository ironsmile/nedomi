package upstream

import (
	"net"

	"github.com/ironsmile/nedomi/types"
)

func (u *Upstream) initDNSResolver(
	algo types.UpstreamBalancingAlgorithm,
	upstreams []*types.UpstreamAddress,
	logger types.Logger,
) {
	//!TODO: use cancel channel
	//!TODO: implement an intelligent TTL-aware persistent resolver
	result := []*types.UpstreamAddress{}

	for _, up := range upstreams {
		ips, err := net.LookupIP(up.Hostname)
		if err != nil {
			logger.Errorf("Ignoring upstream %s: %s", &up.URL, err)
			continue
		}

		for _, ip := range ips {
			if !u.config.Settings.UseIPv4 && ip.To4() != nil {
				continue
			}
			if !u.config.Settings.UseIPv6 && ip.To4() == nil {
				continue
			}

			resolved := *up
			resolved.Hostname = ip.String()
			resolved.Host = net.JoinHostPort(ip.String(), up.Port)
			result = append(result, &resolved)
		}
	}

	algo.Set(result)
	logger.Logf("Finished resolving the upstream IPs for %s; found %d", u.config.ID, len(result))
}
