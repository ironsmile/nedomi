package vhost

import (
	"testing"

	"github.com/ironsmile/nedomi/config"
)

func TestCreatingAVhost(t *testing.T) {
	vh := New(config.VirtualHost{}, nil, nil)
	if vh == nil {
		t.Error("Creating a vhost returned nil")
	}
}
