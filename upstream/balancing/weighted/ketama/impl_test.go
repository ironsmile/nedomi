package ketama

import (
	"net/url"
	"testing"

	"github.com/ironsmile/nedomi/types"
)

type testServer struct {
	host   string
	weight uint32
}
type testCase struct{ path, expected string }

var servers = []testServer{}
var testCases = []testCase{}

func TestForCompliance(t *testing.T) {
	upstreams := make([]*types.UpstreamAddress, len(servers))
	for i, s := range servers {
		upstreams[i] = &types.UpstreamAddress{
			URL:         url.URL{Host: s.host, Scheme: "http"},
			Hostname:    s.host,
			Port:        "80",
			OriginalURL: &url.URL{Host: s.host, Scheme: "http"},
			Weight:      s.weight,
		}
	}
	ketama := New()
	ketama.Set(upstreams)
	for _, test := range testCases {
		if res, err := ketama.Get(test.path); err != nil {
			t.Errorf("Unexpected error when getting %s: %s", test.path, err)
		} else if res.Host != test.expected {
			t.Errorf("Expected to receive %s when getting %s but received %s instead", test.expected, test.path, res.Host)
		}
	}
}
