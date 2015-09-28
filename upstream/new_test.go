package upstream

import (
	"testing"
)

func TestCreatingBogusUpstream(t *testing.T) {
	t.Parallel()
	_, err := New("bogus_upstream", nil)

	if err == nil {
		t.Error("There was no error when creating bogus upstream")
	}
}
