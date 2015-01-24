package config

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func projectPath() (string, error) {
	gopath := os.ExpandEnv("$GOPATH")
	relPath := filepath.FromSlash("src/github.com/gophergala/nedomi")
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

	if err := cfg.Verify(); err != nil {
		t.Errorf("Example config verification had error: %s", err)
	}

}

func getNormalConfig() *Config {
	return &Config{
		Listen: ":5435",
		System: SystemSection{Pidfile: filepath.Join(os.TempDir(), "nedomi.pid")},
	}
}

func TestConfigVerification(t *testing.T) {
	cfg := getNormalConfig()

	if err := cfg.Verify(); err != nil {
		t.Errorf("Got error on working config: %s", err)
	}

	tests := map[string]func(*Config){
		"No error with empty Listen": func(cfg *Config) {
			cfg.Listen = ""
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
		if err := cfg.Verify(); err == nil {
			t.Errorf(errorStr)
		}
	}
}
