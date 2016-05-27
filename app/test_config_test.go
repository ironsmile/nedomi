package app

import (
	"bytes"
	"testing"

	"github.com/ironsmile/nedomi/config"
	"github.com/ironsmile/nedomi/types"
	"github.com/ironsmile/nedomi/utils/testutils"
)

const (
	replaceHolder = `REPLACEME`
)

var testAppVersion = types.AppVersion{}

var configIssues = []struct {
	name   string
	err    string
	config string
}{
	{
		name: "128",
		err:  `No upstream set for proxy handler in 127.0.0.6`,
		config: `{
"system": {
	"pidfile": "/tmp/nedomi_pidfile.pid",
	"workdir": "/tmp/"
},
"default_cache_type": "disk",
"default_cache_algorithm": "lru",
"cache_zones": {
	"default": {
		"path": "REPLACEME",
		"storage_objects": 100,
		"part_size": "2m"
	}
},
"http": {
	"listen":":8282",
	"default_handlers": [{
		"type" : "cache"
	},
	{
		"type" : "proxy"
	}],
	"virtual_hosts": {
		"127.0.0.6": {
			"cache_zone": "zone2",
			"cache_key": "7.4"
		}
	}
},
"logger": {
	"type": "std",
	"settings": {
		"level": "debug"
	}
}}`,
	},
	{
		name: "163",
		err:  `No such balancing algorithm: bogus-algorithm-which-does-not-exist`,
		config: `{
"system": {
	"pidfile": "/tmp/nedomi_pidfile.pid",
	"workdir": "/tmp/"
},
"default_cache_type": "disk",
"default_cache_algorithm": "lru",
"cache_zones": {
	"default": {
		"path": "REPLACEME",
		"storage_objects": 100,
		"part_size": "2m"
	}
},
"http": {
	"upstreams": {
		"nana": {
			"balancing": "bogus-algorithm-which-does-not-exist",
			"servers": ["http://127.0.0.1/", "http://127.0.0.2"]
		}
	},
	"listen":":8282",
	"upstream": "nana",
	"default_handlers": [{
		"type" : "proxy"
	}],
	"virtual_hosts": {
		"127.0.0.6": {
			"upstream": "nana",
			"cache_zone": "zone2",
			"cache_key": "7.4"
		}
	}
},
"logger": {
	"type": "std",
	"settings": {
		"level": "debug"
	}
}}`,
	}, {
		name: "211",
		err:  `error while opening file [/not/existingi/directory/I/hope/it/is/atleast.log] for log output: open /not/existingi/directory/I/hope/it/is/atleast.log: no such file or directory`,
		config: `{
"system": {
	"pidfile": "/tmp/nedomi_pidfile.pid",
	"workdir": "/tmp/"
},
"default_cache_type": "disk",
"default_cache_algorithm": "lru",
"cache_zones": {
	"default": {
		"path": "REPLACEME",
		"storage_objects": 100,
		"part_size": "2m"
	}
},
"http": {
	"listen":":8282",
	"default_handlers": [{
		"type" : "status"
	}],
	"virtual_hosts": {
		"127.0.0.6": {
			"cache_zone": "zone2",
			"cache_key": "7.4"
		}
	}
},
"logger": {
	"type": "ironsmile",
	"settings": {
		"log": "/not/existingi/directory/I/hope/it/is/atleast.log"
	}
}}`,
	}, {
		name: "201",
		err:  `handler must have a type`,
		config: `{
"system": {
	"pidfile": "/tmp/nedomi_pidfile.pid",
	"workdir": "/tmp/"
},
"default_cache_type": "disk",
"default_cache_algorithm": "lru",
"cache_zones": {
	"default": {
		"path": "REPLACEME",
		"storage_objects": 100,
		"part_size": "2m"
	}
},
"http": {
	"listen":":8282",
	"default_handlers": [{
		"type_" : "status"
	}],
	"virtual_hosts": {
		"127.0.0.6": {
			"cache_zone": "zone2",
			"cache_key": "7.4"
		}
	}
},
"logger": {
	"type": "std",
	"settings": {
		"level": "debug"
	}
}}`,
	}, {
		name: "167",
		err:  "Could not initialize storage 'disk' for cache zone 'default': Disk storage path `/hopefully/non/existing/cache/path` should be created.",
		config: `{
"system": {
	"pidfile": "/tmp/nedomi_pidfile.pid",
	"workdir": "/tmp/"
},
"default_cache_type": "disk",
"default_cache_algorithm": "lru",
"cache_zones": {
	"default": {
		"path": "/hopefully/non/existing/cache/path",
		"storage_objects": 100,
		"part_size": "2m"
	}
},
"http": {

	"listen":":8282",
	"default_handlers": [{
		"type" : "status"
	}],
	"virtual_hosts": {
		"127.0.0.6": {
			"cache_zone": "zone2",
			"cache_key": "7.4"
		}
	}
},
"logger": {
	"type": "std",
	"settings": {
		"level": "debug"
	}
}}`,
	}, /*{ // This is not fixed yet
			name: "177",
			err:  `change`,
			config: `{
	"system": {
		"pidfile": "/tmp/nedomi_pidfile.pid",
		"workdir": "/tmp/"
	},
	"default_cache_type": "disk",
	"default_cache_algorithm": "lru",
	"cache_zones": {
		"default": {
			"path": "REPLACEME",
			"storage_objects": 100,
			"part_size": "2m"
		}
	},
	"http": {
		"upstreams": {
			"nana": {
				"balancing": "random",
				"servers": ["http://127.0.0.1/", "http://127.0.0.2"]
			}
		},
		"listen":":8282",
		"default_handlers": [{
			"type" : "proxy",
			"settings": {
				"try_other_upstream_on_code": {
					"404": "does-not-exist"
				}
			}
		}],
		"virtual_hosts": {
			"127.0.0.6": {
				"upstream": "nana",
				"cache_zone": "zone2",
				"cache_key": "7.4"
			}
		}
	},
	"logger": {
		"type": "std",
		"settings": {
			"level": "debug"
		}
	}}`,
		},*/
}

func TestConfigIssues(t *testing.T) {
	t.Parallel()
	for _, issue := range configIssues {

		testConfigGetsError(t, issue.name, issue.config, issue.err)
	}
}

func testConfigGetsError(t *testing.T, name, cfg, expected string) {
	folder, cleanup := testutils.GetTestFolder(t)
	defer cleanup()
	f := func() (*config.Config, error) {
		return config.ParseBytes(replace(cfg, folder))
	}

	_, err := New(testAppVersion, f)
	if err == nil {
		if expected != "" {
			t.Errorf("test for `%s`: expected error `%s`, got nothing", name, expected)
		}
		return
	}

	if err.Error() != expected {
		t.Errorf("test for `%s`: expected error `%s`, got `%s`", name, expected, err)
	}
}

func replace(a, b string) []byte {
	return bytes.Replace([]byte(a), []byte(replaceHolder), []byte(b), 1)
}
