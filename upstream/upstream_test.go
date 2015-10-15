package upstream

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
			input: &url.URL{Host: "127.0.0.1"},
			host:  "127.0.0.1",
			port:  "",
			err:   true,
		},
	}

	for _, test := range tests {
		host, port, err := parseURLHost(test.input)
		if test.err && err == nil {
			t.Errorf("expected error for input '%+v' but didn't get any",
				test.input)
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
