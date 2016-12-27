package headers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/ironsmile/nedomi/config"
)

func testHeaders(t *testing.T, key string, expected []string, headers http.Header) {
	var got = headers[http.CanonicalHeaderKey(key)]
	if !reflect.DeepEqual(got, expected) {
		t.Errorf("for key '%s' expected '%+v', got '%+v'", key, expected, got)
	}

}

func handlerCode(code int) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(code)
	})
}

func addHeaderHandler(t *testing.T, header string, value string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
	}`)), nil, handlerCode(200))

	if err != nil {
		t.Errorf("Got error when initializing via - %s", err)
	}
	var expect []string
	recorder := httptest.NewRecorder()
	req, err := http.NewRequest("get", "/to/test", nil)
	if err != nil {
		t.Fatal(err)
	}
	v.ServeHTTP(recorder, req)
	expect = []string{testText}
	testHeaders(t, canonicalKey, expect, recorder.Header())

	recorder.Header().Set(canonicalKey, "holla")
	expect = []string{"holla", testText}
	v.ServeHTTP(recorder, req)
	testHeaders(t, canonicalKey, expect, recorder.Header())
}

func TestRemoveFromRequest(t *testing.T) {
	t.Parallel()
	var expectedHeaders = map[string][]string{
		"Via":     nil,
		"Added":   {"old value", "added value"},
		"Setting": {"this", "header"},
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
	}`)), nil, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
	v.ServeHTTP(recorder, req)
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
	v.ServeHTTP(recorder, req)
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
	}`)), nil, handlerCode(200))
	if err == nil {
		t.Errorf("Expected error on initializing with broken config but got %+v", v)
	}
}

func TestCodes(t *testing.T) {
	t.Parallel()
	var codes = []int{http.StatusPartialContent, http.StatusOK, http.StatusNotFound, http.StatusTeapot}
	for _, code := range codes {
		canonicalKey := http.CanonicalHeaderKey("via")
		var testText = "notnedomi 2.2"
		v, err := New(config.NewHandler("headers", json.RawMessage(`{
		"response": {
			"add_headers": {
				"vIa": "notnedomi 2.2"
			}
		},
		"request": {
			"add_headers": {
				"via": "notnedomi 2.2"
			}
		}
	}`)), nil, handlerCode(code))

		if err != nil {
			t.Errorf("Got error when initializing via - %s", err)
		}
		var expect []string
		rec := httptest.NewRecorder()
		req, err := http.NewRequest("get", "/to/test", nil)
		if err != nil {
			t.Fatal(err)
		}
		v.ServeHTTP(rec, req)
		expect = []string{testText}
		testHeaders(t, canonicalKey, expect, rec.Header())
		if code != rec.Code {
			t.Errorf("expected code %d got %d", code, rec.Code)
		}
	}
}
