package upstream

import (
	"testing"
)

func TestCreatingBogusUpstream(t *testing.T) {
	_, err := New("bogus_upstream", nil)

	if err == nil {
		t.Error("There was no error when creating bogus upstream")
	}
}
