package app

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
	"sync"

	"github.com/ironsmile/nedomi/types"
	"github.com/tchap/go-patricia/patricia"
)

type locationType string

const (
	none                   locationType = "/"
	exact                               = "= /"
	caseSensitiveRegular                = "~ "
	caseInsensitiveRegular              = "~* "
	bestNonRegular                      = "^~ /"
)

var errExactMatch = fmt.Errorf("exact match") // not an actual error

// LocationMuxer is muxer for Location types. Given a slice of them it can tell which one should respond to a given path.
type LocationMuxer struct {
	sync.Mutex
	regexes      []*regexLocation
	locationTrie *patricia.Trie
}

type regexLocation struct {
	*regexp.Regexp
	*types.Location
}

func isLocationType(location *types.Location, locType locationType) bool {
	return strings.HasPrefix(location.Name, (string)(locType))
}

// NewLocationMuxer returns a new LocationMuxer for the given Locations
func NewLocationMuxer(locations []*types.Location) (*LocationMuxer, error) {
	lm := new(LocationMuxer)
	lm.locationTrie = patricia.NewTrie()
	for _, location := range locations {
		switch {
		case isLocationType(location, none):
			lm.locationTrie.Set([]byte(location.Name), location)
		case isLocationType(location, caseInsensitiveRegular):
			if err := lm.addRegexForLocation(fmt.Sprintf(`(?i)%s`, location.Name[len(caseInsensitiveRegular):]), location); err != nil {
				return nil, fmt.Errorf("Location %s gave error while being parsed to regex: %s", location, err)
			}
		case isLocationType(location, caseSensitiveRegular):
			if err := lm.addRegexForLocation(location.Name[len(caseSensitiveRegular):], location); err != nil {
				return nil, fmt.Errorf("Location %s gave error while being parsed to regex: %s", location, err)
			}
		case isLocationType(location, exact):
			lm.locationTrie.Set([]byte(location.Name[len(exact)-1:]), location)
		case isLocationType(location, bestNonRegular):
			lm.locationTrie.Set([]byte(location.Name[len(bestNonRegular)-1:]), location)
		default:
			return nil, fmt.Errorf("Location %s is not parsable", location)

		}
	}

	return lm, nil
}

func (lm *LocationMuxer) addRegexForLocation(regex string, location *types.Location) error {
	reg, err := regexp.Compile(regex)
	if err != nil {
		return fmt.Errorf("Location %s gave error while being parsed to regex: %s", location, err)
	}
	lm.regexes = append(lm.regexes, &regexLocation{
		Regexp:   reg,
		Location: location,
	})
	return nil
}

func (lm *LocationMuxer) longestMatchingPath(path string) *types.Location {
	pathAsPrefix := patricia.Prefix(path)
	var matchSoFar patricia.Prefix
	var matchedItem *types.Location
	err := lm.locationTrie.Visit(func(prefix patricia.Prefix, item patricia.Item) error {
		if len(prefix) > len(pathAsPrefix) {
			return patricia.SkipSubtree
		} else if len(prefix) > len(matchSoFar) && bytes.EqualFold(prefix, pathAsPrefix[:len(prefix)]) {
			exactMatch := len(prefix) == len(pathAsPrefix)
			matchedLocation := item.(*types.Location)
			if isLocationType(matchedLocation, exact) && !exactMatch {
				return nil
			}

			matchedItem = matchedLocation
			matchSoFar = prefix
			if exactMatch {
				return errExactMatch // not an error, just so the search is canceled
			}
		}
		return nil
	})

	if err != nil && err != errExactMatch {
		panic(err) // an impossible error
	}

	return matchedItem
}

func (lm *LocationMuxer) firstMatchingRegex(path string) *types.Location {
	for _, regex := range lm.regexes {
		if regex.MatchString(path) {
			return regex.Location
		}
	}

	return nil
}

// Match returns the Location which should respond to the given path
func (lm *LocationMuxer) Match(path string) *types.Location {
	lm.Lock()
	defer lm.Unlock()
	location := lm.longestMatchingPath(path)
	if location != nil && isLocationType(location, bestNonRegular) {
		return location
	}

	regexLocation := lm.firstMatchingRegex(path)
	if regexLocation != nil {
		return regexLocation
	}
	return location
}
