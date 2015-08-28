package types

import (
	"strconv"
	"strings"
	"testing"

	"fmt"
)

func TestObjectIndexStringersWithSensibleData(t *testing.T) {
	objIdx := ObjectIndex{
		ObjID: &ObjectID{
			CacheKey: "1.2",
			Path:     "/somewhere",
		},
		Part: 33,
	}

	result := fmt.Sprintf("%s", objIdx)

	if len(result) < 1 {
		t.Error("The stringer for ObjectIndex returned empty string")
	}

	if !strings.Contains(result, strconv.Itoa(int(objIdx.Part))) {
		t.Errorf("The result '%s' does not contain the part number '%d'", result, objIdx.Part)
	}
}
