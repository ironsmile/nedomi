package headers

import (
	"net/http"
	"reflect"
	"testing"

	"github.com/ironsmile/nedomi/config"
)

func TestComplex(t *testing.T) {
	var headers = make(http.Header)
	headers.Add("foo", "2")
	headers.Add("foo", "3")
	headers.Add("bar", "2")
	headers.Add("bar", "3")
	headers.Add("foobar", "2")
	headers.Add("foobar", "3")
	headers.Add("foofoo", "2")
	headers.Add("foofoo", "3")
	headers.Add("barbar", "2")
	headers.Add("barbar", "3")
	var hr = &headersRewrite{
		AddHeaders: config.HeaderPairs{
			"foo":    {"is", "cool"},
			"foobar": {"is", "cool"},
			"foofoo": {"is", "cool"},
		},
		SetHeaders: config.HeaderPairs{
			"foobar": {"not", "cool"},
			"barbar": {"not", "cool"},
		},
		RemoveHeaders: []string{"bar", "foo"},
	}

	hr.rewrite(headers)
	var tests = []struct {
		key      string
		expected []string
	}{
		{key: "Foo", expected: []string{"is", "cool"}},
		{key: "Bar", expected: nil},
		{key: "FooBar", expected: []string{"not", "cool"}},
		{key: "FooFoo", expected: []string{"2", "3", "is", "cool"}},
		{key: "BarBar", expected: []string{"not", "cool"}},
	}
	for _, test := range tests {
		got := headers[http.CanonicalHeaderKey(test.key)]
		if !reflect.DeepEqual(got, test.expected) {
			t.Errorf("for '%s': expected '%+v' got '%+v'", test.key, test.expected, got)
		}
	}
}
