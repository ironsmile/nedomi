package headers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"golang.org/x/net/context"

	"github.com/ironsmile/nedomi/config"
	"github.com/ironsmile/nedomi/types"
)

func testStringHandler(t *testing.T, txt string) types.RequestHandler {
	return types.RequestHandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		var values = w.Header()[http.CanonicalHeaderKey("via")]
		if len(values) == 0 {
			t.Errorf("no value for via key")
		} else if values[len(values)-1] != txt {
			t.Errorf("wrong value for via")
		}
	})
}

func addHeaderHandler(t *testing.T, header string, value string) types.RequestHandler {
	return types.RequestHandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		w.Header().Add(header, value)
		w.WriteHeader(200) // this is actually needed
	})
}

func TestVia(t *testing.T) {
	t.Parallel()
	canonicalKey := http.CanonicalHeaderKey("via")
	var testText = "notnedomi 2.2"
	v, err := New(config.NewHandler("headers", json.RawMessage(`{
		"response": {
			"add_headers": {
				"vIa": "notnedomi 2.2"
			}
		}
	}`)), nil, testStringHandler(t, testText))

	if err != nil {
		t.Errorf("Got error when initializing via - %s", err)
	}
	var expect, got []string
	recorder := httptest.NewRecorder()
	req, err := http.NewRequest("get", "/to/test", nil)
	if err != nil {
		t.Fatal(err)
	}
	v.RequestHandle(nil, recorder, req)
	expect = []string{testText}
	got = recorder.Header()[canonicalKey]
	if !reflect.DeepEqual(got, expect) {
		t.Errorf("expected via header to be equal to %s but got %s", expect, got)
	}

	recorder.Header().Set(canonicalKey, "holla")
	expect = []string{"holla", testText}
	v.RequestHandle(nil, recorder, req)
	got = recorder.Header()[canonicalKey]
	if !reflect.DeepEqual(got, expect) {
		t.Errorf("expected via header to be equal to %s but got %s", expect, got)
	}
}

func TestRemoveFromRequest(t *testing.T) {
	t.Parallel()
	var expectedHeaders = map[string][]string{
		"Via":     nil,
		"Added":   []string{"old value", "added value"},
		"Setting": []string{"this", "header"},
	}
	v, err := New(config.NewHandler("headers", json.RawMessage(`{
		"request": {
			"remove_headers": ["vIa"],
			"add_headers":  {
				"added": "added value"
			},
			"set_headers": {
				"setting": ["this", "header"]
			}
		}
	}`)), nil, types.RequestHandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		for key, expected := range expectedHeaders {
			got := r.Header[http.CanonicalHeaderKey(key)]
			if !reflect.DeepEqual(got, expected) {
				t.Errorf("for header '%s' expected '%+v', got '%+v'", key, got, expected)
			}
		}
	}))

	if err != nil {
		t.Errorf("Got error when initializing via - %s", err)
	}
	recorder := httptest.NewRecorder()
	req, err := http.NewRequest("get", "/to/test", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Add(http.CanonicalHeaderKey("via"), "this should be removed")
	req.Header.Add(http.CanonicalHeaderKey("added"), "old value")
	req.Header.Add(http.CanonicalHeaderKey("setting"), "this should be resetted")
	v.RequestHandle(nil, recorder, req)
}

func TestRemoveFromResponse(t *testing.T) {
	t.Parallel()
	canonicalKey := http.CanonicalHeaderKey("via")
	v, err := New(config.NewHandler("headers", json.RawMessage(`{
		"response": {
			"remove_headers": ["vIa"]
		}
	}`)), nil, addHeaderHandler(t, "via", "pesho"))

	if err != nil {
		t.Errorf("Got error when initializing via - %s", err)
	}
	recorder := httptest.NewRecorder()
	req, err := http.NewRequest("get", "/to/test", nil)
	if err != nil {
		t.Fatal(err)
	}
	v.RequestHandle(nil, recorder, req)
	var got = recorder.Header()[canonicalKey]
	if len(got) != 0 {
		t.Errorf("expected via header to be removed but got %s", got)
	}
}

func TestNilNext(t *testing.T) {
	t.Parallel()
	v, err := New(config.NewHandler("headers", json.RawMessage(`{
		"response": {
			"add_headers": {
				"vIa": "notnedomi 2.2"
			}
		}
	}`)), nil, nil)
	if err == nil {
		t.Errorf("Expected error on initializing with nil next handler but got %+v", v)
	}
}

func TestWithBrokenConfig(t *testing.T) {
	t.Parallel()
	v, err := New(config.NewHandler("headers", json.RawMessage(`{
		"response: {
			"add_headers": {
				"vIa": "notnedomi 2.2"
			}
		}
	}`)), nil, testStringHandler(t, "pesho"))
	if err == nil {
		t.Errorf("Expected error on initializing with broken config but got %+v", v)
	}
}
