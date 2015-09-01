package types

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"fmt"
)

func TestObjectIDHash(t *testing.T) {
	obj := NewObjectID("1.2", "/somewhere")

	res := obj.StrHash()
	expected := "583fae38a17840864d328e08b0d21cec293f74b2"
	if expected != res {
		t.Errorf("Incorrect ObjectID hash. Exected %s, got %s", expected, res)
	}
}

func TestObjectIDJsonHandling(t *testing.T) {
	obj := NewObjectID("1.2", "/somewhere")
	expectedRes := "[\"1.2\",\"/somewhere\"]"
	resM, err := json.Marshal(obj)
	if err != nil {
		t.Fatalf("Could not marshal ObjectID: %s", err)
	}
	if string(resM) != expectedRes {
		t.Fatalf("Result %s differs from the expected result %s", resM, expectedRes)
	}

	resU := &ObjectID{}
	if err := json.Unmarshal(resM, resU); err != nil {
		t.Fatalf("Could not unmarshal ObjectID: %s", err)
	}
	if !reflect.DeepEqual(obj, resU) {
		t.Fatalf("The original object %#v is different from the unmarshalled %#v", obj, resU)
	}
}

func TestObjectIDJsonErrors(t *testing.T) {
	wrongStrings := []string{"", "[]", "{}", "[\"test\"]", "[\"\",\"\"]",
		"[\"test\",\"\"]", "[\"\",\"test\"]", "\"test\""}

	tmp := &ObjectID{}
	for _, v := range wrongStrings {
		if err := json.Unmarshal([]byte(v), tmp); err == nil {
			t.Errorf("Expected to have an error with %s", v)
		}
	}
}

func TestObjectIDStringersWithSensibleData(t *testing.T) {
	obj := NewObjectID("1.2", "/somewhere")
	result := fmt.Sprintf("%s", obj)

	if len(result) < 1 {
		t.Error("The stringer for ObjectID returned empty string")
	}

	if !strings.Contains(result, obj.CacheKey()) {
		t.Errorf("The result '%s' does not contain the cache key '%s'", result, obj.CacheKey())
	}

	if !strings.Contains(result, obj.Path()) {
		t.Errorf("The result '%s' does not contain the path '%s'", result, obj.Path())
	}

	if !strings.Contains(result, obj.StrHash()) {
		t.Errorf("The result '%s' does not contain the hash '%s'", result, obj.StrHash())
	}

}
