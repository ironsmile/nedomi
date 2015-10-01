package app

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/ironsmile/nedomi/types"
)

func newVHost(name string) *VirtualHost {
	muxer, err := NewLocationMuxer(nil)
	if err != nil {
		panic(err)
	}
	return &VirtualHost{
		Location: types.Location{
			Name: name,
		},
		Muxer: muxer,
	}
}

func TestVirtualHostsMaching(t *testing.T) {
	t.Parallel()
	app := &Application{
		virtualHosts: map[string]*VirtualHost{
			"localhost":            newVHost("localhost"),
			"server.com":           newVHost("server.com"),
			"subdomain.server.com": newVHost("subdomain.server.com"),
			"10.8.3.43":            newVHost("10.8.3.43"),
		},
	}

	for reqInd, reqName := range []string{
		"localhost",
		"server.com",
		"subdomain.server.com",
		"10.8.3.43",
	} {
		req := &http.Request{
			Host: fmt.Sprintf("%s:%d", reqName, 80+reqInd),
		}

		location := app.GetLocationFor(req.Host, "")

		if location.Name != reqName {
			t.Errorf("Expected to find location for %s but it found %s", reqName,
				location.Name)
		}
	}

	location := app.GetLocationFor("no-such-host-here.com:993", "")

	if location != nil {
		t.Errorf("Searching for non matching location returned one: %s",
			location.Name)
	}

}
