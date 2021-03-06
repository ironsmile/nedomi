package config

import (
	"testing"
	"time"
)

func TestLocationJSONUnmarshallingValidLocation(t *testing.T) {
	t.Parallel()
	loc := newLocForTesting()

	err := loc.UnmarshalJSON([]byte(`
		{
			"upstream": "upstream1",
			"cache_zone": "default",
			"handlers": [{"type": "status"}]
		}
	`))

	if err != nil {
		t.Errorf("Error while json unmrashalling working location: %s", err)
	}
}

var defaultDurtaionMatrix = []struct {
	sectionString  string
	unmarshallable bool
	verifyOK       bool
	duration       time.Duration
}{
	{
		sectionString: `
			{
				"upstream": "upstream1",
				"handlers": [{"type": "cache"}],
				"cache_zone": "default",
				"cache_key": "1.1",
				"cache_default_duration": "3h",
				"cache_key_includes_query": true
			}
		`,
		unmarshallable: true,
		verifyOK:       true,
		duration:       3 * time.Hour,
	}, {
		sectionString: `
			{
				"upstream": "upstream1",
				"handlers": [{"type": "cache"}],
				"cache_zone": "default",
				"cache_key": "1.1",
				"cache_default_duration": "baba",
				"cache_key_includes_query": true
			}
		`,
		unmarshallable: false,
		verifyOK:       false,
	}, {
		sectionString: `
			{
				"upstream": "upstream1",
				"handlers": [{"type": "cache"}],
				"cache_zone": "default",
				"cache_key": "1.1",
				"cache_key_includes_query": true
			}
		`,
		unmarshallable: true,
		verifyOK:       true,
		duration:       2 * DefaultCacheDuration, //!TODO: change this when this is configurable
	}, {
		sectionString: `
			{
				"upstream": "upstream1",
				"handlers": [{"type": "cache"}],
				"cache_zone": "default",
				"cache_key": "1.1",
				"cache_default_duration": "-5h",
				"cache_key_includes_query": true
			}
		`,
		unmarshallable: true,
		verifyOK:       false,
	},
}

func TestLocationJSONUnmarshallingAndVeirfyingCacheDefaultDuration(t *testing.T) {
	t.Parallel()
	for index, test := range defaultDurtaionMatrix {
		loc := newLocForTesting()
		err := loc.UnmarshalJSON([]byte(test.sectionString))
		if test.unmarshallable && err != nil {
			t.Errorf("Error while unmarshalling working config %d: %s", index, err)
		} else if !test.unmarshallable && err == nil {
			t.Errorf("No error while unmarshalling broken working config %d", index)
		}

		if !test.unmarshallable {
			continue
		}

		err = loc.Validate()

		if test.verifyOK && err != nil {
			t.Errorf("Error while verifying working config %d: %s", index, err)
		}

		if !test.verifyOK && err == nil {
			t.Errorf("No error while verifying broken config %d: %s", index, err)
		}

		if !test.verifyOK {
			continue
		}

		if test.duration != loc.CacheDefaultDuration {
			t.Errorf("Expected default cache duration of %s but got: %s",
				test.duration, loc.CacheDefaultDuration)
		}
	}
}

func newLocForTesting() *Location {
	loc := new(Location)
	cfg := &Config{
		CacheZones: make(map[string]*CacheZone),
	}

	loc.Name = "/baba"
	loc.parent = &VirtualHost{}
	loc.parent.parent = &HTTP{}
	loc.parent.parent.parent = cfg

	// Potentionally a panic
	cfg.CacheZones["default"] = nil

	//!TODO: This is hardcoded to the same value as the one in the JSON Unmarshalling in
	// the vhost section.
	loc.parent.CacheDefaultDuration = 2 * DefaultCacheDuration

	return loc
}
