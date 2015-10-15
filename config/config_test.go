package config

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/ironsmile/nedomi/utils/testutils"
)

//!TODO: split this in multiple files and write additional tests

func TestExampleConfig(t *testing.T) {
	t.Parallel()
	path, err := testutils.ProjectPath()

	if err != nil {
		t.Fatalf("Was not able to find project path: %s", err)
	}

	if _, err := Parse("not-present-config.json"); err == nil {
		t.Errorf("Expected error when parsing non existing config but got nil")
	}

	examplePath := filepath.Join(path, "config.example.json")
	cfg, err := Parse(examplePath)
	if err != nil {
		t.Fatalf("Parsing the example config returned: %s", err)
	}

	if err := cfg.Validate(); err != nil {
		t.Errorf("Example config verification had error: %s", err)
	}

}

func getNormalConfig() *Config {
	//!TODO: split into different test case composable builders
	c := new(Config)
	c.System = System{Pidfile: filepath.Join(os.TempDir(), "nedomi.pid")}
	c.Logger = *NewLogger("nillogger", nil)

	cz := &CacheZone{
		ID:             "test1",
		Type:           "disk",
		Path:           os.TempDir(),
		StorageObjects: 20,
		PartSize:       1024,
		Algorithm:      "lru",
	}
	c.CacheZones = map[string]*CacheZone{"test1": cz}

	loc := Location{
		baseLocation: baseLocation{
			Name:     "localhost",
			CacheKey: "test",
			Handlers: []Handler{*NewHandler("proxy", nil)},
			Logger:   c.Logger,
		},
		CacheZone:            cz,
		CacheDefaultDuration: 2 * time.Hour,
	}

	c.HTTP = new(HTTP)
	c.HTTP.Listen = ":5435"
	c.HTTP.Logger = c.Logger
	c.HTTP.Servers = []*VirtualHost{{
		baseVirtualHost: baseVirtualHost{
			Aliases: []string{
				"localhost-1", "localhost-2",
			},
		},
		Location: loc,
		Locations: []*Location{
			&loc,
		},
	}}
	c.HTTP.Upstreams = []*Upstream{{
		ID:        "test1",
		Balancing: "single",
		Addresses: []UpstreamAddress{
			{URL: &url.URL{Scheme: "http", Host: "www.google.com"}},
		},
	}}
	c.HTTP.Servers[0].Upstream = "test1"

	return c
}

func TestConfigVerification(t *testing.T) {
	t.Parallel()
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
		"No error with wrong cache default duration in location": func(cfg *Config) {
			cfg.HTTP.Servers[0].Locations[0].CacheDefaultDuration = -1 * time.Hour
		},
		"No error with wrong cache default duration in vhost": func(cfg *Config) {
			cfg.HTTP.Servers[0].CacheDefaultDuration = -1 * time.Hour
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

func TestHandlersParsing(t *testing.T) {
	t.Parallel()
	cfg, err := parseBytes([]byte(`
{
	"system": {
		"pidfile": "/tmp/nedomi_pidfile.pid",
		"workdir": "/tmp/"
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
		"default_handlers": [
		{ "type" : "via" },
		{ "type" : "proxy", "settings" : {"field" : "proxySetting"}}
		],
		"default_upstream_type": "simple",
		"default_cache_zone": "default",
		"virtual_hosts": {
			"localhost": {
				"upstream": "http://upstream.com",
				"cache_key": "1.1",
				"locations": {
					"/status": {
						"handlers": [
						{ "type" : "gzip"},
						{ "type" : "via", "settings": {"field": "viasetting"}},
						{ "type" : "status"}
						]
					},
					"~ \\.mp4$": {
						"logger": {
							"type": "notLogger",
							"settings" : {"field" : "notField"}
						}
					},
					"~* \\.mp4$": {
						"handlers": [ { "type" : "proxy"}]
					}

				},
				"logger": {
					"type": "localhostLogger",
					"settings" : {"field" : "loalField"}
				}
			},
			"127.0.0.2": {
				"handlers": [{ "type" : "status" }]
			}
		},
		"logger": {
			"type": "httplogger",
			"settings" : {"field" : "nilLogger"}
		}
	},
	"logger": {
		"type": "nillogger",
		"settings" : {"field" : "nilLogger"}
	}
}
`))
	//!TODO do it with http section only

	if err != nil {
		t.Fatalf("Got error parsing config - %s", err)
	}

	// don't touch it works
	var mat struct {
		Def     []DefStruct `json:"def"`
		Logger  DefStruct   `json:"logger"`
		Servers map[string]struct {
			Def       []DefStruct `json:"def"`
			Logger    DefStruct   `json:"logger"`
			Locations map[string]struct {
				Handlers []DefStruct `json:"handlers"`
				Logger   DefStruct   `json:"logger"`
			} `json:"locations"`
		} `json:"servers"`
	}
	matText := []byte(`
	{
		"def": [
		{ "type": "via"} ,
		{ "type" : "proxy", "setting": "proxySetting"} ],
		"logger" : {"type" : "httplogger", "setting" : "nilLogger"},
		"servers": {
			"localhost": {
				"def": [{ "type": "via"}, {"type": "proxy", "setting" : "proxySetting"}],
				"logger": {"type": "localhostLogger", "setting": "loalField"},
				"locations": {
					"/status": {
						"handlers":[ {"type" :"gzip"}, {"type" : "via", "setting": "viasetting"}, {"type" : "status"}],
						"logger": {"type": "localhostLogger", "setting": "loalField"}
					},
					"~ \\.mp4$": {
						"handlers" :[{"type" :"via"},  {"type": "proxy", "setting": "proxySetting"}],
						"logger": {"type": "notLogger", "setting": "notField"}
					},
					"~* \\.mp4$": {
						"handlers" : [ {"type":"proxy"}],
						"logger": {"type": "localhostLogger", "setting": "loalField"}
					}
				}
			},
			"127.0.0.2": {
				"def": [ {"type":"status"} ],
				"logger" : {"type" : "httplogger", "setting" : "nilLogger"}
			}
		}
	}`)
	err = json.Unmarshal(matText, &mat)
	if err != nil {
		t.Fatalf("Got error parsing config - %s", err)
	}

	checkHandlerTypes(t, "Default Handlers are wrong", cfg.HTTP.DefaultHandlers, mat.Def)
	checkLogger(t, "Default Logger is wrong", cfg.HTTP.Logger, mat.Logger)
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
		checkLogger(t, fmt.Sprintf("Default logger for server %s is wrong", server), server.Logger, serverExpect.Logger)
		for _, location := range server.Locations {
			locationExpect, ok := serverExpect.Locations[location.Name]
			if !ok {
				t.Errorf("got a location %s that is not expected", location)
				continue
			}
			checkHandlerTypes(t, fmt.Sprintf("Handlers for location %s are wrong", location), location.Handlers, locationExpect.Handlers)
			checkLogger(t, fmt.Sprintf("Logger for location %s is wrong", location), location.Logger, locationExpect.Logger)
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

func checkLogger(t *testing.T, msg string, logger Logger, loggerExpect DefStruct) {
	if loggerExpect.Type != logger.Type {
		t.Errorf("%s expected `%s`, got `%s`", msg, loggerExpect.Type, logger.Type)
		return
	}
	var s struct {
		Field string `json:"field"`
	}
	if len(logger.Settings) != 0 {
		err := json.Unmarshal(logger.Settings, &s)
		if err != nil {
			t.Errorf("got error while parsing Settings for logger %s with raw settings `%s` - %s", logger.Type, string(logger.Settings), err)
		}
		if s.Field != loggerExpect.Setting {
			t.Errorf("%s expected to have Setting `%s`, got `%s`", msg, loggerExpect.Setting, s.Field)
		}
	} else if len(loggerExpect.Setting) > 0 {
		t.Errorf("%s expected to have Setting `%s`, but no Setting was found", msg, loggerExpect.Setting)
	}
}

func checkHandlerTypes(t *testing.T, msg string, handlers []Handler, handlerTypes []DefStruct) {
	if len(handlers) != len(handlerTypes) {
		t.Errorf("%s expected `%s`, got `%s`", msg, handlerTypes, handlers)
		return
	}
	for index := range handlerTypes {
		if handlers[index].Type != handlerTypes[index].Type {
			t.Errorf("%s expected `%s`, got `%s`", msg, handlerTypes[index].Type, handlers[index].Type)
			return
		}
		var s struct {
			Field string `json:"field"`
		}
		if len(handlers[index].Settings) != 0 {
			err := json.Unmarshal(handlers[index].Settings, &s)
			if err != nil {
				t.Errorf("got error while parsing Settings for handler %s with raw settings `%s` - %s", handlers[index].Type, string(handlers[index].Settings), err)
			}
			if s.Field != handlerTypes[index].Setting {
				t.Errorf("%s expected to have Setting `%s`, got `%s`", msg, handlerTypes[index].Setting, s.Field)
			}
		} else if len(handlerTypes[index].Setting) > 0 {
			t.Errorf("%s expected to have Setting `%s`, but no Setting was found", msg, handlerTypes[index].Setting)
		}
	}
}

type DefStruct struct {
	Type    string `json:"type"`
	Setting string `json:"setting"`
}
