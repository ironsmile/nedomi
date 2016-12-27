package testutils

import (
	"fmt"
	"math/rand"
	"testing"
)

func TestGenerateMeAString(t *testing.T) {
	t.Parallel()
	var tests = []struct {
		seed, size int64
	}{
		{seed: 1, size: 10},
		{seed: 2, size: 10},
		{seed: 1, size: 11},
		{seed: 2, size: 11},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("Seed %d, size %d", test.seed, test.size), func(t *testing.T) {
			s := GenerateMeAString(test.seed, test.size)
			if len(s) != int(test.size) {
				t.Errorf("expected `%s` to have size %d not %d", s, test.size, len(s))
			}
		})
	}
}

func TestGetRandomUpstreams(t *testing.T) {
	t.Parallel()
	for i := 0; i < 100; i++ {
		min := rand.Intn(i*17 + 1)
		max := rand.Intn(i*43+1) + min + 1
		upstreams := GetRandomUpstreams(min, max)
		if len(upstreams) < min || len(upstreams) > max {
			t.Errorf("Expected between %d and %d  upstreams got %d", min, max, len(upstreams))
		}

	}
}
