package upstream

import (
	"fmt"
	"net"
	"net/url"
	"strings"

	"github.com/ironsmile/nedomi/types"
)

func parseURLHost(u *url.URL) (host, port string, err error) {
	if strings.ContainsRune(u.Host, ':') && !strings.HasSuffix(u.Host, "]") {
		return net.SplitHostPort(u.Host)
	}

	if u.Scheme == "http" {
		return net.SplitHostPort(u.Host + ":80")
	} else if u.Scheme == "https" {
		return net.SplitHostPort(u.Host + ":443")
	}

	return u.Host, "", fmt.Errorf("address %s has an invalid scheme", u)
}

func (u *Upstream) initDNSResolver(algo types.UpstreamBalancingAlgorithm) {
	//!TODO: use cancel channel
	//!TODO: implement an intelligent TTL-aware persistent resolver
	result := []*types.UpstreamAddress{}

	for _, addr := range u.config.Addresses {
		host, port, err := parseURLHost(addr.URL)
		if err != nil {
			u.logger.Errorf("Ignoring upstream %s: %s", addr.URL, err)
			continue
		}

		ips, err := net.LookupIP(host)
		if err != nil {
			u.logger.Errorf("Ignoring upstream %s: %s", addr.URL, err)
			continue
		}

		for _, ip := range ips {
			if !u.config.Settings.UseIPv4 && ip.To4() != nil {
				continue
			}
			if !u.config.Settings.UseIPv6 && ip.To4() == nil {
				continue
			}

			resolved := *addr.URL
			resolved.Host = net.JoinHostPort(ip.String(), port)
			result = append(result, &types.UpstreamAddress{
				URL:         addr.URL,
				ResolvedURL: &resolved,
				Weight:      addr.Weight,
			})
			algo.Set(result)
		}
	}

	u.logger.Logf("Finished resolving the upstream IPs for %s; found %d", u.config.ID, len(result))
}
