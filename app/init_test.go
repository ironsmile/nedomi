package app

import (
	"net/http"
	"path/filepath"
	"testing"

	"github.com/ironsmile/nedomi/config"
	"github.com/ironsmile/nedomi/utils/testutils"
)

func TestConcurrentCacheReload(t *testing.T) {
	t.Parallel()
	t.Skip("TODO: implement")
}

func TestAliasesMatchingAfterInit(t *testing.T) {
	t.Parallel()

	path, err := testutils.ProjectPath()

	if err != nil {
		t.Fatalf("Was not able to find the project dir: %s", err)
	}

	//!TODO: maybe construct an config ourselves
	// We are using the example config for this test. This might not be
	// so great an idea. But I tried to construct a config programatically
	// for about an hour and a half and I failed.
	examplePath := filepath.Join(path, "config.example.json")
	cfg, err := config.Parse(examplePath)

	if err != nil {
		t.Fatalf("Error parsing example config: %s", err)
	}

	// To make sure no output is emitted during testing
	cfg.Logger.Type = "nillogger"

	app, err := New(cfg)

	if err != nil {
		t.Fatalf("Error creating an app: %s", err)
	}

	if err := app.initFromConfig(); err != nil {
		t.Fatalf("Error initializing app: %s", err)
	}

	expected := app.findVirtualHost(&http.Request{Host: "127.0.0.2"})
	found := app.findVirtualHost(&http.Request{Host: "127.0.1.2"})

	if expected != found {
		t.Errorf("Expected vhost %s but got %s", expected, found)
	}

}
