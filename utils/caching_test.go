package utils

import (
	"bufio"
	"bytes"
	"io"
	"net/http"
	"net/textproto"
	"testing"
	"time"
)

var responseCachabilityMatrix = []struct {
	// the code of the response
	code int
	// the string represenation of the headers
	headers string
	// whether this should be cacheable
	cacheable bool
	// for how much it should be cacheable
	expiresIN time.Duration
	// how much of deviation expiresIN could have
	//and the test could still pass
	slack time.Duration
}{
	{
		code:      http.StatusOK,
		headers:   `Cache-Control: private`,
		cacheable: false,
	},
	{
		code:      http.StatusOK,
		headers:   `Cache-Control: max-age`,
		cacheable: false,
	},
	{
		code:      http.StatusOK,
		headers:   `Cache-Control: max-age=30`,
		cacheable: true,
		expiresIN: time.Second * 30,
	},
	{
		code:      http.StatusOK,
		headers:   `Cache-Control: max-age=30, s-maxage=20`,
		cacheable: true,
		expiresIN: time.Second * 20,
	},
	{
		code:      http.StatusOK,
		headers:   `Cache-Control: no-cache`,
		cacheable: false, // actually true with expiresIN of 0 but no plans for supporting it yet
	},
	{
		code:      http.StatusOK,
		headers:   `Cache-Control: no-store, max-age=30`,
		cacheable: false,
	},
	{
		code:      http.StatusOK,
		headers:   `Expires: ` + time.Now().Add(30*time.Second).Format(time.RFC1123),
		cacheable: true,
		expiresIN: time.Second * 30,
		slack:     time.Second,
	},
	{
		code:      http.StatusOK,
		headers:   "Cache-Control: max-age=40\nExpires: " + time.Now().Add(30*time.Second).Format(time.RFC1123),
		cacheable: true,
		expiresIN: time.Second * 40,
	},
	{
		code:      http.StatusOK,
		headers:   "Cache-Control: private\nExpires: " + time.Now().Add(30*time.Second).Format(time.RFC1123),
		cacheable: false,
	},
	{
		code:      http.StatusTeapot, // Teapots are not cacheable
		headers:   `Expires: ` + time.Now().Add(30*time.Second).Format(time.RFC1123),
		cacheable: false,
	},

	{
		code:      http.StatusOK,
		headers:   "Content-Encoding: tea\nExpires: " + time.Now().Add(30*time.Second).Format(time.RFC1123),
		cacheable: false,
	},
}

func TestIsResponseCachable(t *testing.T) {
	for index, test := range responseCachabilityMatrix {
		headers, err := textproto.NewReader(bufio.NewReader(bytes.NewReader([]byte(test.headers)))).ReadMIMEHeader()
		if err != nil && err != io.EOF {
			t.Errorf("got error %s while parsing headers for test with index %d and headers:\n%s", err, index, test.headers)
		}
		cacheable, expiresIn := IsResponseCacheable(test.code, http.Header(headers))
		if cacheable != test.cacheable {
			if cacheable {
				t.Errorf("NOT cacheable response was said to be cacheable at index %d : `\n%+v`", index, test)
			} else {
				t.Errorf("cacheable response was said to be NOT cacheable at index %d : `\n%+v`", index, test)
			}
		}
		if test.expiresIN-test.slack > expiresIn || expiresIn > test.expiresIN+test.slack {
			t.Errorf("for index %d the expected expired is %s but got %s : \n%+v", index, test.expiresIN, expiresIn, test)
		}
	}
}
