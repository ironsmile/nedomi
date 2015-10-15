package httputils

import "net/http"

// Error is short for http.Error(w, http.StatusText(code), code)
func Error(w http.ResponseWriter, code int) {
	http.Error(w, http.StatusText(code), code)
}
