package roundrobin

import (
	"math"
	"reflect"
	"testing"

	"github.com/ironsmile/nedomi/types"
)

func TestRoundRobin(t *testing.T) {
	t.Parallel()

	rr := New()
	if _, err := rr.Get("test"); err == nil {
		t.Error("Expected get with no upstreams to return an error")
	}

	compareExpected := func(exp *types.UpstreamAddress) {
		res, err := rr.Get("somepath")
		if err != nil {
			t.Errorf("Received an unexpected error: %s", err)
		} else if !reflect.DeepEqual(exp, res) {
			t.Errorf("Expected to receive %#v but received %#v", exp, res)
		}
	}

	h1 := &types.UpstreamAddress{Hostname: "host1"}
	h2 := &types.UpstreamAddress{Hostname: "host2"}
	rr.Set([]*types.UpstreamAddress{h1})
	compareExpected(h1)
	compareExpected(h1)
	compareExpected(h1)

	rr.Set([]*types.UpstreamAddress{h1, h2})
	compareExpected(h1)
	compareExpected(h2)
	compareExpected(h1)
	compareExpected(h2)

	rr.Set([]*types.UpstreamAddress{h1, h2})
	compareExpected(h1)
	compareExpected(h2)

	// Overflow counter
	rr.counter = math.MaxUint32
	compareExpected(h2)
	compareExpected(h1)
}
