package testutils

import "testing"

func TestGenerateMeAString(t *testing.T) {
	var tests = []struct {
		seed, size int64
	}{
		{seed: 1, size: 10},
		{seed: 2, size: 10},
		{seed: 1, size: 11},
		{seed: 2, size: 11},
	}

	for _, test := range tests {
		s := GenerateMeAString(test.seed, test.size)
		if len(s) != int(test.size) {
			t.Errorf("expected `%s` to have size %d not %d", s, test.size, len(s))
		}
	}
}
