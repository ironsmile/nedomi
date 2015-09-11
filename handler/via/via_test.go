package via

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

func TestVia(t *testing.T) {
	canonicalKey := http.CanonicalHeaderKey("via")
	var testText = "notnedomi 2.2"
	v, err := New(config.NewHandler("via", json.RawMessage(`{"text": "notnedomi 2.2"}`)), nil,
		testStringHandler(t, testText))

	if err != nil {
		t.Errorf("Got error when initializing via - %s", err)
	}
	var expect, got []string
	recorder := httptest.NewRecorder()
	v.RequestHandle(nil, recorder, nil, nil)
	expect = []string{testText}
	got = recorder.Header()[canonicalKey]
	if !reflect.DeepEqual(got, expect) {
		t.Errorf("expected via header to be equal to %s but got %s", expect, got)
	}

	recorder.Header().Set(canonicalKey, "holla")
	expect = []string{"holla", testText}
	v.RequestHandle(nil, recorder, nil, nil)
	got = recorder.Header()[canonicalKey]
	if !reflect.DeepEqual(got, expect) {
		t.Errorf("expected via header to be equal to %s but got %s", expect, got)
	}
}

func testStringHandler(t *testing.T, txt string) types.RequestHandler {
	return types.RequestHandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request, l *types.Location) {
		var values = w.Header()[http.CanonicalHeaderKey("via")]
		if values[len(values)-1] != txt {
			t.Errorf("wrong value for via")
		}
	})
}
