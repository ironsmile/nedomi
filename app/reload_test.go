package app

import (
	"testing"

	"github.com/ironsmile/nedomi/config"
)

func TestCacheZonesAreCompatible(t *testing.T) {
	var tests = []struct {
		cfg1, cfg2 map[string]*config.CacheZone
		err        string
	}{
		{}, // nil's are compatible
		{ // this will blow up in other places
			cfg1: map[string]*config.CacheZone{
				"pesho": {
					ID: "pesho",
				},
			},
		},
		{ // no configs at first, more configs later
			cfg2: map[string]*config.CacheZone{
				"pesho": {
					ID: "pesho",
				},
			},
		},
		{ // same config
			cfg1: map[string]*config.CacheZone{
				"pesho": {
					ID: "pesho",
				},
			},
			cfg2: map[string]*config.CacheZone{
				"pesho": {
					ID: "pesho",
				},
			},
		},
		{ // different types
			cfg1: map[string]*config.CacheZone{
				"pesho": {
					ID:   "pesho",
					Type: "type1",
				},
			},
			cfg2: map[string]*config.CacheZone{
				"pesho": {
					ID:   "pesho",
					Type: "type2",
				},
			},
			err: "different types for same id 'pesho' between configs",
		},
		{ // different paths
			cfg1: map[string]*config.CacheZone{
				"pesho": {
					ID:   "pesho",
					Type: "type1",
					Path: "/path/to/nowhere",
				},
			},
			cfg2: map[string]*config.CacheZone{
				"pesho": {
					ID:   "pesho",
					Type: "type1",
					Path: "/path/to/somewhere",
				},
			},
			err: "different paths for same id 'pesho' between configs",
		},
		{ // different algorithms
			cfg1: map[string]*config.CacheZone{
				"pesho": {
					ID:        "pesho",
					Type:      "type1",
					Path:      "/path/to/somewhere",
					Algorithm: "algorithm1",
				},
			},
			cfg2: map[string]*config.CacheZone{
				"pesho": {
					ID:        "pesho",
					Type:      "type1",
					Path:      "/path/to/somewhere",
					Algorithm: "algorithm2",
				},
			},
			err: "different algorithms for same id 'pesho' between configs",
		},
		{ // different part size
			cfg1: map[string]*config.CacheZone{
				"pesho": {
					ID:        "pesho",
					Type:      "type1",
					Path:      "/path/to/somewhere",
					Algorithm: "algorithm",
					PartSize:  10,
				},
			},
			cfg2: map[string]*config.CacheZone{
				"pesho": {
					ID:        "pesho",
					Type:      "type1",
					Path:      "/path/to/somewhere",
					Algorithm: "algorithm",
					PartSize:  20,
				},
			},
			err: "different part size for same id 'pesho' between configs",
		},
		{ // object size going up is fine
			cfg1: map[string]*config.CacheZone{
				"pesho": {
					ID:             "pesho",
					Type:           "type1",
					Path:           "/path/to/somewhere",
					Algorithm:      "algorithm",
					PartSize:       10,
					StorageObjects: 500,
				},
			},
			cfg2: map[string]*config.CacheZone{
				"pesho": {
					ID:             "pesho",
					Type:           "type1",
					Path:           "/path/to/somewhere",
					Algorithm:      "algorithm",
					PartSize:       10,
					StorageObjects: 600,
				},
			},
		},
		{ // object size going down is fine
			cfg1: map[string]*config.CacheZone{
				"pesho": {
					ID:             "pesho",
					Type:           "type1",
					Path:           "/path/to/somewhere",
					Algorithm:      "algorithm",
					PartSize:       10,
					StorageObjects: 500,
				},
			},
			cfg2: map[string]*config.CacheZone{
				"pesho": {
					ID:             "pesho",
					Type:           "type1",
					Path:           "/path/to/somewhere",
					Algorithm:      "algorithm",
					PartSize:       10,
					StorageObjects: 400,
				},
			},
		},
	}

	for _, test := range tests {
		err := cacheZonesAreCompatible(test.cfg1, test.cfg2)
		if (err == nil && test.err != "") || (err != nil && err.Error() != test.err) {
			t.Errorf("Comparing \n`%+v`\n and \n`%+v`\n  returned `%s` expected `%s`",
				test.cfg1, test.cfg2, err, test.err)
		}
	}
}

func TestCheckConfigCouldBeReloaded(t *testing.T) {
	var tests = []struct {
		cfg1, cfg2 *config.Config
		err        error
	}{{
		cfg1: &config.Config{
			BaseConfig: config.BaseConfig{
				System: config.System{
					Pidfile: "/path/to/pid/file",
					Workdir: "/path/to/work/dir",
					User:    "user",
				},
			},
			HTTP: &config.HTTP{
				BaseHTTP: config.BaseHTTP{
					Listen: ":8282",
				},
			},
		},
		cfg2: &config.Config{
			BaseConfig: config.BaseConfig{
				System: config.System{
					Pidfile: "/path/to/pid/file",
					Workdir: "/path/to/work/dir",
					User:    "user",
				},
			},
			HTTP: &config.HTTP{
				BaseHTTP: config.BaseHTTP{
					Listen: ":8282",
				},
			},
		},
	}, {
		cfg1: &config.Config{
			BaseConfig: config.BaseConfig{
				System: config.System{
					Pidfile: "/path/to/pid/file",
					Workdir: "/path/to/work/dir",
					User:    "user",
				},
			},
			HTTP: &config.HTTP{
				BaseHTTP: config.BaseHTTP{
					Listen: ":8282",
				},
			},
		},
		cfg2: &config.Config{
			BaseConfig: config.BaseConfig{
				System: config.System{
					Pidfile: "/path/to/pid/file/different",
					Workdir: "/path/to/work/dir",
					User:    "user",
				},
			},
			HTTP: &config.HTTP{
				BaseHTTP: config.BaseHTTP{
					Listen: ":8282",
				},
			},
		},
	}, {
		cfg1: &config.Config{
			BaseConfig: config.BaseConfig{
				System: config.System{
					Pidfile: "/path/to/pid/file",
					Workdir: "/path/to/work/dir",
					User:    "user",
				},
			},
			HTTP: &config.HTTP{
				BaseHTTP: config.BaseHTTP{
					Listen: ":8282",
				},
			},
		},
		cfg2: &config.Config{
			BaseConfig: config.BaseConfig{
				System: config.System{
					Pidfile: "/path/to/pid/file",
					Workdir: "/path/to/work/dir/different",
					User:    "user",
				},
			},
			HTTP: &config.HTTP{
				BaseHTTP: config.BaseHTTP{
					Listen: ":8282",
				},
			},
		},
		err: errCfgWorkDirIsDifferent,
	}, {
		cfg1: &config.Config{
			BaseConfig: config.BaseConfig{
				System: config.System{
					Pidfile: "/path/to/pid/file",
					Workdir: "/path/to/work/dir",
					User:    "user",
				},
			},
			HTTP: &config.HTTP{
				BaseHTTP: config.BaseHTTP{
					Listen: ":8282",
				},
			},
		},
		cfg2: &config.Config{
			BaseConfig: config.BaseConfig{
				System: config.System{
					Pidfile: "/path/to/pid/file",
					Workdir: "/path/to/work/dir",
					User:    "user2",
				},
			},
			HTTP: &config.HTTP{
				BaseHTTP: config.BaseHTTP{
					Listen: ":8282",
				},
			},
		},
		err: errCfgUserIsDifferent,
	}, {
		cfg1: &config.Config{
			BaseConfig: config.BaseConfig{
				System: config.System{
					Pidfile: "/path/to/pid/file",
					Workdir: "/path/to/work/dir",
					User:    "user",
				},
			},
			HTTP: &config.HTTP{
				BaseHTTP: config.BaseHTTP{
					Listen: ":8282",
				},
			},
		},
		cfg2: &config.Config{
			BaseConfig: config.BaseConfig{
				System: config.System{
					Pidfile: "/path/to/pid/file",
					Workdir: "/path/to/work/dir",
					User:    "user",
				},
			},
			HTTP: &config.HTTP{
				BaseHTTP: config.BaseHTTP{
					Listen: ":8281",
				},
			},
		},
		err: errCfgListenIsDifferent,
	}, {
		cfg1: &config.Config{
			BaseConfig: config.BaseConfig{
				System: config.System{
					Pidfile: "/path/to/pid/file",
					Workdir: "/path/to/work/dir",
					User:    "user",
				},
			},
			HTTP: &config.HTTP{
				BaseHTTP: config.BaseHTTP{
					Listen: ":8282",
				},
			},
		},
		err: errCfgIsNil,
	}}

	for _, test := range tests {
		a := new(Application)
		a.cfg = test.cfg1
		err := a.checkConfigCouldBeReloaded(test.cfg2)
		if err != test.err {
			t.Errorf("check reloading for  \n`%+v`\n and \n`%+v`\n  returned `%s` expected `%s`",
				test.cfg1, test.cfg2, err, test.err)
		}
	}
}
