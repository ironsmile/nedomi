package utils

import "net/http"

// CopyHeadersWithout copies headers from `from` to `to` except for the `exceptions`
func CopyHeadersWithout(from, to http.Header, exceptions ...string) {
	for k := range from {
		shouldCopy := true
		for _, e := range exceptions {
			if e == k {
				shouldCopy = false
				break
			}
		}
		if shouldCopy {
			to[k] = from[k]
		}
	}
}
