package lru

import (
	"testing"
)

func TestStatsPercentsStringRepresentation(t *testing.T) {
	stats := LruCacheStats{
		id:       "/nana",
		hits:     15,
		requests: 100,
		size:     7322,
		objects:  23,
	}

	found := stats.CacheHitPrc()
	expected := "15%"

	if found != expected {
		t.Errorf("Calculating percents failed. Expected %s but got %s", expected, found)
	}

	stats.requests = 0

	found = stats.CacheHitPrc()

	if found != "" {
		t.Errorf("Calculating percents when no requests returned %s", found)
	}

	stats.requests = 107

	found = stats.CacheHitPrc()
	expected = "14%"

	if found != expected {
		t.Errorf("Calculating percents failed. Expected %s but got %s", expected, found)
	}
}
