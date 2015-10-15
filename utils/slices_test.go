package utils

import (
	"reflect"
	"testing"
)

func TestCopyStringSlice(t *testing.T) {
	var r = []string{"1", "2", "3"}
	var l = CopyStringSlice(r)
	if !reflect.DeepEqual(l, r) {
		t.Errorf("expected '%+v' and '%+v' to be deeply equal", l, r)
	}
	l[1] = "4"
	if reflect.DeepEqual(l, r) {
		t.Errorf("expected '%+v' and '%+v' to not be deeply equal", l, r)
	}
}
