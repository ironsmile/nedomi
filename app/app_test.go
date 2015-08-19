package app

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/ironsmile/nedomi/types"
)

func TestVirtualHostsMaching(t *testing.T) {
	app := &Application{
		virtualHosts: map[string]*types.VirtualHost{
			"localhost": &types.VirtualHost{
				Name: "localhost",
			},
			"server.com": &types.VirtualHost{
				Name: "server.com",
			},
			"subdomain.server.com": &types.VirtualHost{
				Name: "subdomain.server.com",
			},
			"10.8.3.43": &types.VirtualHost{
				Name: "10.8.3.43",
			},
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

		foundVhost := app.findVirtualHost(req)

		if foundVhost.Name != reqName {
			t.Errorf("Expected to find vhost for %s but it found %s", reqName,
				foundVhost.Name)
		}
	}

	foundVhost := app.findVirtualHost(&http.Request{
		Host: "no-such-host-here.com:993",
	})

	if foundVhost != nil {
		t.Errorf("Searching for non existing virtual host returned one: %s",
			foundVhost.Name)
	}

}
