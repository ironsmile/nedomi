package types

import (
	"testing"

	"fmt"
)

func TestStringersWithSensibleData(t *testing.T) {
	obj := &ObjectID{
		CacheKey: "1.2",
		Path:     "/somewhere",
	}

	if result := fmt.Sprintf("%s", obj); len(result) < 1 {
		t.Error("The stringer for ObjectID returned empty string")
	}

	objID := &ObjectIndex{
		ObjID: *obj,
		Part:  33,
	}

	if result := fmt.Sprintf("%s", objID); len(result) < 1 {
		t.Error("The stringer for ObjectIndex returned empty string")
	}

}
