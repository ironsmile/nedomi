package app

import (
	"path/filepath"
	"testing"

	"github.com/ironsmile/nedomi/config"
	"github.com/ironsmile/nedomi/types"
	"github.com/ironsmile/nedomi/utils"
	"github.com/ironsmile/nedomi/utils/testutils"
)

func TestConcurrentCacheReload(t *testing.T) {
	t.Parallel()
	t.Skip("TODO: implement")
}

func configGetterForExampleConfig(path, defaultPath, zone2Path string) config.Getter {
	return func() (*config.Config, error) {
		var examplePath = filepath.Join(path, "config.example.json")
		var cfg, err = config.Parse(examplePath)
		if err != nil {
			return nil, err
		}

		// Create temporary direcotories for the cache zones
		cfg.CacheZones["default"].Path = defaultPath
		cfg.CacheZones["zone2"].Path = zone2Path

		// To make sure no output is emitted during testing
		cfg.Logger.Type = "nillogger"
		return cfg, nil
	}
}

func appFromExampleConfig(t *testing.T) (*Application, func()) {
	var path, err = utils.ProjectPath()
	if err != nil {
		t.Fatalf("Was not able to find the project dir: %s", err)
	}

	path1, cleanup1 := testutils.GetTestFolder(t)
	path2, cleanup2 := testutils.GetTestFolder(t)

	//!TODO: maybe construct an config ourselves
	// We are using the example config for this test. This might not be
	// so great an idea. But I tried to construct a config programmatically
	// for about an hour and a half and I failed.
	var configGetter = configGetterForExampleConfig(path, path1, path2)
	app, err := New(types.AppVersion{}, configGetter)
	if err != nil {
		t.Fatalf("Error creating an app: %s", err)
	}

	if err := app.reinitFromConfig(app.cfg, false); err != nil {
		t.Fatalf("Error initializing app: %s", err)
	}

	return app, func() {
		cleanup1()
		cleanup2()
	}
}

func TestAliasesMatchingAfterInit(t *testing.T) {
	t.Parallel()

	app, cleanup := appFromExampleConfig(t)
	defer cleanup()
	expected := app.GetLocationFor("127.0.0.2", "")
	found := app.GetLocationFor("127.0.1.2", "")

	if expected != found {
		t.Errorf("Expected location %s but got %s", expected, found)
	}
}

func TestReinit(t *testing.T) {
	t.Parallel()

	app, cleanup := appFromExampleConfig(t)
	defer cleanup()
	cfg := *app.cfg
	path3, cleanup3 := testutils.GetTestFolder(t)
	defer cleanup3()
	cfg.CacheZones["zone3"] = &config.CacheZone{
		ID:             "zone3",
		Type:           "disk",
		Path:           path3,
		StorageObjects: 300,
		PartSize:       4096,
		Algorithm:      "lru",
	}
	replaceZone(&cfg, "zone2", cfg.CacheZones["zone3"])
	if err := app.reinitFromConfig(&cfg, false); err != nil {
		t.Fatalf("Error upon reiniting app: %s", err)
	}

	if _, ok := app.cacheZones["zone3"]; !ok {
		t.Error("No zone3 cache zone after reinit")
	}

	if _, ok := app.cacheZones["zone2"]; ok {
		t.Error("zone2 cache zone still present after reinit")
	}
}

func replaceZone(cfg *config.Config, id string, newZone *config.CacheZone) {
	delete(cfg.CacheZones, id)
	for _, server := range cfg.HTTP.Servers {
		if server.CacheZone.ID == id {
			server.CacheZone = newZone
		}
		for _, location := range server.Locations {
			if location.CacheZone.ID == id {
				location.CacheZone = newZone
			}
		}
	}
}
