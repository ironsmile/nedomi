package types

import (
	"strings"
	"testing"

	"fmt"
)

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

}
