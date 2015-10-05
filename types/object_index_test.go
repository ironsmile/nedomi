package types

import (
	"strconv"
	"strings"
	"testing"

	"fmt"
)

func TestObjectIndexStringersWithSensibleData(t *testing.T) {
	t.Parallel()
	objIdx := &ObjectIndex{
		ObjID: NewObjectID("1.2", "/somewhere"),
		Part:  33,
	}

	result := fmt.Sprintf("%s", objIdx)

	if len(result) < 1 {
		t.Error("The stringer for ObjectIndex returned empty string")
	}

	if !strings.Contains(result, strconv.Itoa(int(objIdx.Part))) {
		t.Errorf("The result '%s' does not contain the part number '%d'", result, objIdx.Part)
	}
}

func TestObjectIndexHashe(t *testing.T) {
	t.Parallel()
	objID := NewObjectID("1.3", "/else")
	objIdx1 := &ObjectIndex{ObjID: objID, Part: 0}

	var expectedHash ObjectIndexHash
	copy(expectedHash[:], objID.hash[:])

	if expectedHash != objIdx1.Hash() {
		t.Errorf("Expected hash1 to be '%x' but got '%x'", expectedHash, objIdx1.Hash())
	}

	objIdx2 := &ObjectIndex{ObjID: objID, Part: 4}
	expectedHash[ObjectIndexHashSize-1] = 4
	if expectedHash != objIdx2.Hash() {
		t.Errorf("Expected hash2 to be '%x' but got '%x'", expectedHash, objIdx2.Hash())
	}
}
