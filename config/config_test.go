package config

import (
	"encoding/json"
	"errors"
	"fmt"
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
				Handlers:     []Handler{Handler{Type: "proxy"}},
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

func TestHandlersParsing(t *testing.T) {
	cfg, err := parseBytes([]byte(`
{
    "system": {
        "pidfile": "/tmp/nedomi_pidfile.pid",
        "workdir": "/tmp/test/"
    },
    "default_cache_type": "disk",
    "default_cache_algorithm": "lru",
    "cache_zones": {
        "default": {
            "path": "/tmp/test/cache1",
            "storage_objects": 1023123,
            "part_size": "2m"
        }
    },
    "http": {
        "listen": ":8282",
        "max_headers_size": 1231241212,
        "read_timeout": 12312310,
        "write_timeout": 213412314,
        "default_handler": [
	{ "type" : "via" },
	{ "type" : "proxy" }
	],
        "default_upstream_type": "simple",
        "default_cache_zone": "default",
        "virtual_hosts": {
            "localhost": {
                "upstream_address": "http://upstream.com/",
                "cache_key": "1.1",
                "locations": {
                    "/status": {
                        "handler": [
			{ "type" : "gzip"},
			{ "type" : "via"},
			{ "type" : "status"}
		]
                    },
                    "~ \\.mp4$": {
		    },
                    "~* \\.mp4$": {
			    "handler": [ { "type" : "proxy"}]
                    }
                }
            },
	    "127.0.0.2": {
		    "handler": [{ "type" : "status" }]
	    }
        }
    },
    "logger": {
        "type": "nillogger"
    }
}
`))
	//!TODO do it with http section only

	if err != nil {
		t.Fatalf("Got error parsing config - %s", err)
	}

	// don't touch it works
	var mat struct {
		Def     []string `json:"def"`
		Servers map[string]struct {
			Def       []string            `json:"def"`
			Locations map[string][]string `json:"locations"`
		} `json:"servers"`
	}
	matText := []byte(`
	{
		"def": ["via", "proxy"],
		"servers": {
			"localhost": {
				"def": ["via", "proxy"],
				"locations": {
					"/status": ["gzip", "via", "status"],
					"~ \\.mp4$": ["via", "proxy"],
					"~* \\.mp4$": [ "proxy"]
				}
			},
			"127.0.0.2": {
				"def": ["status" ]
			}
		}
	}`)
	err = json.Unmarshal(matText, &mat)
	if err != nil {
		t.Fatalf("Got error parsing config - %s", err)
	}

	checkHandlerTypes(t, "Default Handlers are wrong", cfg.HTTP.DefaultHandlers, mat.Def)
	if len(cfg.HTTP.Servers) != len(mat.Servers) {
		t.Errorf("expected %d server cfgs got %d", len(mat.Servers), len(cfg.HTTP.Servers))
	}
	for _, server := range cfg.HTTP.Servers {
		serverExpect, ok := mat.Servers[server.Name]
		if !ok {
			t.Errorf("got a server with name %s that is not expected", server.Name)
			continue
		}
		checkHandlerTypes(t, fmt.Sprintf("Default Handlers for server %s are wrong", server), server.Handlers, serverExpect.Def)
		for _, location := range server.Locations {
			locationExpect, ok := serverExpect.Locations[location.Name]
			if !ok {
				t.Errorf("got a location %s that is not expected", location)
				continue
			}
			checkHandlerTypes(t, fmt.Sprintf("Handlers for location %s are wrong", location), location.Handlers, locationExpect)

			delete(serverExpect.Locations, location.Name)
		}
		if len(serverExpect.Locations) != 0 {
			t.Errorf("some locations were not matched %s", serverExpect.Locations)
		}

		delete(mat.Servers, server.Name)
	}
	if len(mat.Servers) != 0 {
		t.Errorf("some servers were not matched %s", mat.Servers)
	}
}

func checkHandlerTypes(t *testing.T, msg string, handlers []Handler, handlerTypes []string) {
	if len(handlers) != len(handlerTypes) {
		t.Errorf("%s expected `%s`, got `%s`", msg, handlerTypes, handlers)
		return
	}
	for index := range handlerTypes {
		if handlers[index].Type != handlerTypes[index] {
			t.Errorf("%s expected `%s`, got `%s`", msg, handlerTypes, handlers)
			return
		}
	}
}
