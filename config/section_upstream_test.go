package config

import (
	"encoding/json"
	"fmt"
	"math"
	"net/url"
	"reflect"
	"testing"
)

type upstreamTestCase struct {
	json             string
	expRes           Upstream
	expValidateError bool
}

var upstreams = []upstreamTestCase{
	{json: `{"balancing":"test","addresses":[]}`, expValidateError: true},
	{json: `{"balancing":"test","addresses":["http://upstream1.com|60%","https://upstream2.com|40.1%"]}`, expValidateError: true},
	{
		json: `{"balancing":"mest","addresses":["http://upstream1.com"],"settings":{"max_connections_per_server":15}}`,
		expRes: Upstream{Balancing: "mest", Addresses: []UpstreamAddress{
			{URL: &url.URL{Scheme: "http", Host: "upstream1.com"}},
		}, Settings: UpstreamSettings{MaxConnectionsPerServer: 15}},
	},
	{
		json: `{"balancing":"test","addresses":["http://user:pass@upstream2.com|20%","https://upstream3.com|33.12%"]}`,
		expRes: Upstream{Balancing: "test", Addresses: []UpstreamAddress{
			{URL: &url.URL{Scheme: "http", Host: "upstream2.com", User: url.UserPassword("user", "pass")}, Weight: 0.2},
			{URL: &url.URL{Scheme: "https", Host: "upstream3.com"}, Weight: 0.3312},
		}, Settings: UpstreamSettings{MaxConnectionsPerServer: 0}},
	},
}

var wrongUpstreams = []string{
	`{wrong_json}`,
	`{"addresses":[""]}`,
	`{"addresses":["http://wrong%url.com"]}`,
	`{"addresses":["http://upstream.com|"]}`,
	`{"addresses":["http://upstream.com|50"]}`,
	`{"addresses":["http://upstream.com|-50%"]}`,
	`{"addresses":["http://upstream.com|baba%"]}`,
	`{"addresses":["http://upstream.com|50.1.2%"]}`,
}

func compareAddresses(exp, res []UpstreamAddress) error {
	if len(exp) != len(res) {
		return fmt.Errorf("expected %#v, received %#v", exp, res)
	}

	for addrNum, addr := range res {
		expAddr := exp[addrNum]
		if addr.URL.String() != expAddr.URL.String() {
			return fmt.Errorf("address %s is different than expected %s", addr.URL, expAddr.URL)
		}
		if math.Abs(addr.Weight-expAddr.Weight) > math.Nextafter(1.0, 2.0)-1.0 {
			return fmt.Errorf("address weight %f for address %s is different than expected %f",
				addr.Weight, addr.URL, expAddr.Weight)
		}
	}
	return nil
}

func TestUpstreamParsingAndValidation(t *testing.T) {
	t.Parallel()

	for testNum, testCase := range upstreams {
		u := &Upstream{}
		if err := json.Unmarshal([]byte(testCase.json), u); err != nil {
			t.Errorf("Unexpected error while parsing upstream %d: %s", testNum, err)
		} else if testCase.expValidateError {
			if u.Validate() == nil {
				t.Errorf("Expected to get validation error for upstream %d", testNum)
			}
		} else if u.Balancing != testCase.expRes.Balancing {
			t.Errorf("Upstream %d had balancing '%s' while expecting '%s'", testNum, u.Balancing, testCase.expRes.Balancing)
		} else if !reflect.DeepEqual(u.Settings, testCase.expRes.Settings) {
			t.Errorf("Upstream %d had settings '%#v' while expecting '%#v'", testNum, u.Settings, testCase.expRes.Settings)
		} else if err := compareAddresses(u.Addresses, testCase.expRes.Addresses); err != nil {
			t.Errorf("Different addresses for upstream %d: %s", testNum, err)
		}
	}
}

func TestWrongUpstreamParsing(t *testing.T) {
	t.Parallel()

	for num, testCase := range wrongUpstreams {
		u := &Upstream{ID: "Test"}
		if err := json.Unmarshal([]byte(testCase), u); err == nil {
			t.Errorf("Expected to receive an error while parsing wrong upstream config %d: %s", num, testCase)
		}
	}

}
