package app

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/ironsmile/nedomi/config"
	"github.com/ironsmile/nedomi/mock"
	"github.com/ironsmile/nedomi/storage/disk"
	"github.com/ironsmile/nedomi/types"
	"github.com/ironsmile/nedomi/utils"
	"github.com/ironsmile/nedomi/utils/testutils"
)

func getConfigGetter(tmpPath string) func() (*config.Config, error) {
	return func() (*config.Config, error) {
		path, err := utils.ProjectPath()
		if err != nil {
			return nil, fmt.Errorf("Was not able to find project path: %s", err)
		}

		cfg, err := config.Parse(filepath.Join(path, "config.example.json"))
		if err != nil {
			return nil, fmt.Errorf("Parsing the example config returned: %s", err)
		}

		for k := range cfg.CacheZones {
			// Fix and create storage paths
			cfg.CacheZones[k].Path = filepath.Join(tmpPath, k)
			if err := os.Mkdir(cfg.CacheZones[k].Path, os.FileMode(0700|os.ModeDir)); err != nil {
				return nil, err
			}
		}

		return cfg, nil
	}
}

func TestDiskReload(t *testing.T) {
	t.Parallel()
	tempDir, cleanup := testutils.GetTestFolder(t)
	defer cleanup()

	app, err := New(types.AppVersion{}, getConfigGetter(tempDir))
	if err != nil {
		t.Fatalf("Could not create an application: %s", err)
	}
	app.SetLogger(mock.NewLogger())

	stor, err := disk.New(app.cfg.CacheZones["default"], app.GetLogger())
	if err != nil {
		t.Fatalf("Could not initialize a storage: %s", err)
	}

	objIDNew := types.NewObjectID("key", "new")
	objIDOld := types.NewObjectID("key", "old")
	testutils.ShouldntFail(t,
		stor.SaveMetadata(&types.ObjectMetadata{ID: objIDNew, ExpiresAt: time.Now().Unix() + 600}),
		stor.SaveMetadata(&types.ObjectMetadata{ID: objIDOld, ExpiresAt: time.Now().Unix() - 600}),
		stor.SavePart(&types.ObjectIndex{ObjID: objIDNew, Part: 0}, strings.NewReader("test1-1")),
		stor.SavePart(&types.ObjectIndex{ObjID: objIDNew, Part: 1}, strings.NewReader("test1-2")),
		stor.SavePart(&types.ObjectIndex{ObjID: objIDOld, Part: 0}, strings.NewReader("test2-1")),
	)

	if err := app.initFromConfig(); err != nil {
		t.Fatalf("Could not init from config: %s", err)
	}
	defer app.ctxCancel()
	time.Sleep(1 * time.Second)

	const expectedObjects = 2
	cacheObjects := app.cacheZones["default"].Algorithm.Stats().Objects()
	if cacheObjects != expectedObjects {
		t.Errorf("Expected object count in cache to be %d but it was %d", expectedObjects, cacheObjects)
	}
}
