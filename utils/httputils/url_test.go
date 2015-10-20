package httputils

import (
	"net/url"
	"testing"
)

func TestParseURLHost(t *testing.T) {
	var tests = []struct {
		input      *url.URL
		host, port string
		err        bool
	}{
		{
			input: &url.URL{Host: "127.0.0.1:8282"},
			host:  "127.0.0.1",
			port:  "8282",
		},
		{
			input: &url.URL{Host: "127.0.0.1", Scheme: "http"},
			host:  "127.0.0.1",
			port:  "80",
		},
		{
			input: &url.URL{Host: "127.0.0.1", Scheme: "https"},
			host:  "127.0.0.1",
			port:  "443",
		},
		{
			input: &url.URL{Host: "ucdn.com:80", Scheme: "https"},
			host:  "ucdn.com",
			port:  "80",
		},
		{
			input: &url.URL{Host: "[2010:836B:4179::836B:4179]:85", Scheme: "https"},
			host:  "2010:836B:4179::836B:4179",
			port:  "85",
		},
		{
			input: &url.URL{Host: "[FEDC:BA98:7654:3210:FEDC:BA98:7654:3210]", Scheme: "https"},
			host:  "FEDC:BA98:7654:3210:FEDC:BA98:7654:3210",
			port:  "443",
		},
		{
			input: &url.URL{Host: "[::192.9.5.5]"},
			err:   true,
		},
		{
			input: &url.URL{Host: "[3ffe:2a00:100:7031::1]", Scheme: "wtf"},
			err:   true,
		},
		{
			input: &url.URL{Host: "127.0.0.1"},
			err:   true,
		},
	}

	for _, test := range tests {
		host, port, err := ParseURLHost(test.input)
		if test.err {
			if err == nil {
				t.Errorf("expected error for input '%+v' but didn't get any",
					test.input)
			}
			continue
		} else if !test.err && err != nil {
			t.Errorf("expected no error for input '%+v' but got '%s'",
				test.input, err)

		}
		if host != test.host {
			t.Errorf("for input '%+v' expected host  '%s' got '%s'",
				test.input, test.host, host)
		}
		if port != test.port {
			t.Errorf("for input '%+v' expected port '%s' got '%s'",
				test.input, test.port, port)
		}
	}

}
