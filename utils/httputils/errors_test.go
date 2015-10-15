package httputils

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestError(t *testing.T) {
	var tests = []int{
		http.StatusMethodNotAllowed,
		http.StatusBadRequest,
		http.StatusInternalServerError,
	}

	for _, code := range tests {
		var rec = httptest.NewRecorder()
		Error(rec, code)
		if rec.Code != code {
			t.Errorf("expected code %d got %d", code, rec.Code)
		}
		var expectedText = http.StatusText(code) + "\n"
		var text = rec.Body.String()
		if text != expectedText {
			t.Errorf("expected body text '%s' got '%s'", expectedText, text)
		}
	}
}
