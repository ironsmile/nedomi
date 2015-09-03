package types

import (
	"testing"
)

func newLocation(match string) *Location {
	return &Location{
		Match: match,
	}
}

var (
	rootLocation           = newLocation("/")
	exactRootLocation      = newLocation("= /")
	bestNonRegularLocation = newLocation("^~ /pictures")
	jpgRegexLocation       = newLocation(`~ jpg$`)
	jpgRegexILocation      = newLocation(`~* jpg$`)
)

var matrix = []struct {
	// make muxer with this locations
	locations []*Location
	// who asked for the key will return the value
	// if the map is nil it means the muxer constructor should return error
	results map[string]*Location
}{
	{
		locations: []*Location{rootLocation},
		results: map[string]*Location{
			"/index.html": rootLocation,
		},
	},
	{
		locations: []*Location{exactRootLocation},
		results: map[string]*Location{
			"/index.html": nil,
			"/":           exactRootLocation,
		},
	},
	{
		locations: []*Location{newLocation("")},
	},
	{
		locations: []*Location{newLocation("notstartging with root")},
	},
	{
		locations: []*Location{newLocation("= notstarting with slash")},
	},
	{
		locations: []*Location{newLocation("^~ not starting with slash")},
	},
	{
		locations: []*Location{newLocation("^~ not starting with slash")},
	},
	{
		locations: []*Location{jpgRegexLocation},
		results: map[string]*Location{
			"/":                         nil,
			"/index.html":               nil,
			"/somewhere/else/test.jpg":  jpgRegexLocation,
			"/somewhere/else/test.Jpg":  nil,
			"/somewhere/else/test.jpge": nil,
		},
	},
	{
		locations: []*Location{jpgRegexLocation, jpgRegexILocation},
		results: map[string]*Location{
			"/":                         nil,
			"/index.html":               nil,
			"/somewhere/else/test.jpg":  jpgRegexLocation,
			"/somewhere/else/test.Jpg":  jpgRegexILocation,
			"/somewhere/else/test.jpge": nil,
		},
	},
	{
		locations: []*Location{exactRootLocation, jpgRegexLocation, jpgRegexILocation, bestNonRegularLocation},
		results: map[string]*Location{
			"/":                         exactRootLocation,
			"/index.html":               nil,
			"/pictures/else/test.jpg":   bestNonRegularLocation,
			"/pictures/else/test.Jpg":   bestNonRegularLocation,
			"/pictures/else/test.jpge":  bestNonRegularLocation,
			"/picturees/else/test.jpg":  jpgRegexLocation,
			"/picturees/else/test.Jpg":  jpgRegexILocation,
			"/picturees/else/test.jpge": nil,
		},
	},
}

func TestMat(t *testing.T) {
	for _, test := range matrix {
		locations := test.locations
		results := test.results
		muxer, err := NewLocationMuxer(locations)
		if err != nil && results != nil {
			t.Errorf("Got error during init - %s for locations:\n%s", err, locations)
			continue
		} else if results == nil && err == nil {
			t.Errorf("Didn't get error while parsing locations:\n%s", locations)
			continue
		}

		for url, location := range results {
			match := muxer.Match(url)

			if match != location {
				t.Errorf("'%s' should've matched location `%s` not `%s`, given locations:\n%s", url, location, match, locations)
			}
		}
	}
}
