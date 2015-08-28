package types

import (
	"strings"
	"testing"

	"fmt"
)

func TestObjectIDHash(t *testing.T) {
	obj := &ObjectID{
		CacheKey: "1.2",
		Path:     "/somewhere",
	}

	res := obj.StrHash()
	expected := "583fae38a17840864d328e08b0d21cec293f74b2"
	if expected != res {
		t.Errorf("Incorrect ObjectID hash. Exected %s, got %s", expected, res)
	}
}

func TestObjectIDStringersWithSensibleData(t *testing.T) {
	obj := &ObjectID{
		CacheKey: "1.2",
		Path:     "/somewhere",
	}

	result := fmt.Sprintf("%s", obj)

	if len(result) < 1 {
		t.Error("The stringer for ObjectID returned empty string")
	}

	if !strings.Contains(result, obj.CacheKey) {
		t.Errorf("The result '%s' does not contain the cache key '%s'", result, obj.CacheKey)
	}

	if !strings.Contains(result, obj.Path) {
		t.Errorf("The result '%s' does not contain the path '%s'", result, obj.Path)
	}

	if !strings.Contains(result, obj.StrHash()) {
		t.Errorf("The result '%s' does not contain the hash '%s'", result, obj.StrHash())
	}

}
