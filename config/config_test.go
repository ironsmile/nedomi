package config

import (
	"errors"
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

	cfg := &Config{}
	examplePath := filepath.Join(path, "config.example.json")

	if err := cfg.Parse(examplePath); err != nil {
		t.Errorf("Parsing the example config returned: %s", err)
	}

	if err := cfg.Parse("not-present-config.json"); err == nil {
		t.Errorf("Expected error when parsing non existing config but got nil")
	}

	if err := cfg.Validate(); err != nil {
		t.Errorf("Example config verification had error: %s", err)
	}

}

func getNormalConfig() *Config {
	c := new(Config)
	c.HTTP = new(HTTP)
	c.DefaultCacheAlgorithm = "lru"
	c.HTTP.Listen = ":5435"
	c.HTTP.DefaultUpstreamType = "simple"
	c.System = SystemSection{Pidfile: filepath.Join(os.TempDir(), "nedomi.pid")}
	c.Logger = LoggerSection{Type: "nillogger"}
	return c
}

func TestConfigVerification(t *testing.T) {
	cfg := getNormalConfig()

	if err := cfg.Validate(); err != nil {
		t.Errorf("Got error on working config: %s", err)
	}

	tests := map[string]func(*Config){
		//!TODO: enable
		//"No error with empty Listen": func(cfg *Config) {
		//	cfg.HTTP.Listen = ""
		//},
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
		if err := cfg.Validate(); err == nil {
			t.Errorf(errorStr)
		}
	}
}
