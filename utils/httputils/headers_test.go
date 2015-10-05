package httputils

import (
	"net/http"
	"reflect"
	"testing"
)

func TestCopyHeaders(t *testing.T) {
	t.Parallel()
	from := http.Header{
		"test": []string{"mest", "pest"},
		"jest": []string{"with", "zest"},
	}
	to := http.Header{}

	CopyHeaders(from, to)
	if !reflect.DeepEqual(from, to) {
		t.Errorf("From and to are different: %#v, %#v", from, to)
	}

	from["jest"][0] = "with no"
	if reflect.DeepEqual(from, to) {
		t.Error("From and to should be different after changing from")
	}
}

func TestCopyHeadersWithout(t *testing.T) {
	t.Parallel()
	from := http.Header{
		"test": []string{"mest", "pest"},
		"jest": []string{"with", "zest"},
		"fest": []string{"best"},
	}
	to := http.Header{"wazzup": []string{"wazzup"}}
	exp := http.Header{"fest": []string{"best"}, "wazzup": []string{"wazzup"}}

	CopyHeadersWithout(from, to, "test", "jest")
	from["fest"][0] = "cancelled"
	if !reflect.DeepEqual(to, exp) {
		t.Errorf("Expected %#v but got %#v", exp, to)
	}
}
