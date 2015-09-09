package config

import (
	"errors"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

//!TODO: split this in multiple files and write additional tests

func projectPath() (string, error) {
	gopath := os.ExpandEnv("$GOPATH")
	relPath := filepath.FromSlash("src/github.com/ironsmile/nedomi")
	for _, path := range strings.Split(gopath, ":") {
		rootPath := filepath.Join(path, relPath)
		entry, err := os.Stat(rootPath)
		if err != nil {
			continue
		}

		if entry.IsDir() {
			return rootPath, nil
		}
	}

	return "", errors.New("Was not able to find the project path")
}

func TestExampleConfig(t *testing.T) {
	path, err := projectPath()

	if err != nil {
		t.Fatalf("Was not able to find project path: %s", err)
	}

	if _, err := parse("not-present-config.json"); err == nil {
		t.Errorf("Expected error when parsing non existing config but got nil")
	}

	examplePath := filepath.Join(path, "config.example.json")
	cfg, err := parse(examplePath)
	if err != nil {
		t.Errorf("Parsing the example config returned: %s", err)
	}

	if err := cfg.Validate(); err != nil {
		t.Errorf("Example config verification had error: %s", err)
	}

}

func getNormalConfig() *Config {
	//!TODO: split into different test case composable builders
	c := new(Config)
	c.System = System{Pidfile: filepath.Join(os.TempDir(), "nedomi.pid")}
	c.Logger = Logger{Type: "nillogger"}

	cz := &CacheZone{
		ID:             "test1",
		Type:           "disk",
		Path:           os.TempDir(),
		StorageObjects: 20,
		PartSize:       1024,
		Algorithm:      "lru",
	}
	c.CacheZones = map[string]*CacheZone{"test1": cz}

	c.HTTP = new(HTTP)
	c.HTTP.Listen = ":5435"
	c.HTTP.Logger = c.Logger
	c.HTTP.Servers = []*VirtualHost{&VirtualHost{
		Location: Location{
			baseLocation: baseLocation{
				Name:         "localhost",
				UpstreamType: "simple",
				CacheKey:     "test",
				Handler:      Handler{Type: "proxy"},
				Logger:       &c.Logger,
			},
			CacheZone: cz,
		},
	}}
	c.HTTP.Servers[0].UpstreamAddress, _ = url.Parse("http://www.google.com")

	return c
}

func TestConfigVerification(t *testing.T) {
	cfg := getNormalConfig()

	if err := ValidateRecursive(cfg); err != nil {
		t.Errorf("Got error on working config: %s", err)
	}

	tests := map[string]func(*Config){
		"No error with empty Listen": func(cfg *Config) {
			cfg.HTTP.Listen = ""
		},
		"No error with wrong pidfile directory": func(cfg *Config) {
			cfg.System.Pidfile = "/does-not-exists/pidfile.pid"
		},
		"No error with wrong empty pidfile": func(cfg *Config) {
			cfg.System.Pidfile = ""
		},
		"No error with wrong user directive": func(cfg *Config) {
			cfg.System.User = "no-existing-user-please"
		},
	}

	for errorStr, fnc := range tests {
		cfg = getNormalConfig()
		fnc(cfg)
		if err := ValidateRecursive(cfg); err == nil {
			t.Errorf(errorStr)
		}
	}
}

func TestDuplicateCacheSettings(t *testing.T) {
	cfg := getNormalConfig()

	if err := ValidateRecursive(cfg); err != nil {
		t.Errorf("Did not expect an error in the normal config")
	}

	cfg.HTTP.Servers = append(cfg.HTTP.Servers, cfg.HTTP.Servers[0])

	if err := ValidateRecursive(cfg); err == nil {
		t.Errorf("Expected an error when having duplicate vhost cache zones and keys")
	}

}
