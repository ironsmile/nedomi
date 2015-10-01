package types

import (
	"net/url"
	"testing"
)

func TestNewObjectIDForURL(t *testing.T) {
	var locations = []*Location{
		&Location{
			CacheKey: "1",
		},
		&Location{
			CacheKey:              "2",
			CacheKeyIncludesQuery: true,
		},
	}
	var tests = map[string][]string{ // url -> []ObjectID.Path
		"/test/path/to/awesome":                    []string{"/test/path/to/awesome", "/test/path/to/awesome"},
		"/test/path/to/awesome?epic=2":             []string{"/test/path/to/awesome", "/test/path/to/awesome?epic=2"},
		"/test/path/to/awesome?epic=2#moreAwesome": []string{"/test/path/to/awesome", "/test/path/to/awesome?epic=2#moreAwesome"},
		"/test/path/to/awesome#moreAwesome":        []string{"/test/path/to/awesome", "/test/path/to/awesome#moreAwesome"},
	}

	for uString, expectations := range tests {
		u, err := url.Parse(uString)
		if err != nil {
			t.Fatal(err)
		}
		for i, l := range locations {
			got := l.NewObjectIDForURL(u)
			expected := expectations[i]
			if got.Path() != expected {
				t.Errorf("expected '%s' got '%s' for url '%s' with location %+v ",
					expected, got.Path(), uString, l)
			}
			if got.CacheKey() != l.CacheKey {
				t.Errorf("expected objectID '%+v' to have the same CacheKey as location %+v ",
					got, l)
			}
		}
	}

}
