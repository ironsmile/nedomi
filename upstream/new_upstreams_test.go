package upstream

import (
	"testing"

	"github.com/ironsmile/nedomi/config"
)

func TestCreatingBogusUpstream(t *testing.T) {
	_, err := New("bogus_upstream", &config.Config{})

	if err == nil {
		t.Error("There was no error when creating bogus upstream")
	}
}
