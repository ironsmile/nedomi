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
	v, err := New(&config.Handler{
		Type: "via",
		Settings: json.RawMessage(`
		{"text": "notnedomi 2.2"}
		`),
	}, nil, types.RequestHandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request, l *types.Location) {
		values := w.Header()[http.CanonicalHeaderKey("via")]
		if values[len(values)-1] != "notnedomi 2.2" {
			t.Errorf("wrong value for via")
		}
	}))

	if err != nil {
		t.Errorf("Got error when initializing via - %s", err)
	}
	var expect, got []string
	recorder := httptest.NewRecorder()
	v.RequestHandle(nil, recorder, nil, nil)
	expect = []string{"notnedomi 2.2"}
	got = recorder.Header()[canonicalKey]
	if !reflect.DeepEqual(got, expect) {
		t.Errorf("expected via header to be equal to %s but got %s", expect, got)
	}

	recorder.Header().Set(canonicalKey, "holla")
	expect = []string{"holla", "notnedomi 2.2"}
	v.RequestHandle(nil, recorder, nil, nil)
	got = recorder.Header()[canonicalKey]
	if !reflect.DeepEqual(got, expect) {
		t.Errorf("expected via header to be equal to %s but got %s", expect, got)
	}
}
