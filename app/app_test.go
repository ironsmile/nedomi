package app

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/ironsmile/nedomi/config"
	"github.com/ironsmile/nedomi/vhost"
)

func TestVirtualHostsMaching(t *testing.T) {
	app := &Application{
		virtualHosts: map[string]*vhostPair{
			"localhost": &vhostPair{
				vhostStruct: vhost.New(config.VirtualHost{
					BaseVirtualHost: config.BaseVirtualHost{Name: "localhost"},
				}, nil, nil),
			},
			"server.com": &vhostPair{
				vhostStruct: vhost.New(config.VirtualHost{
					BaseVirtualHost: config.BaseVirtualHost{Name: "server.com"},
				}, nil, nil),
			},
			"subdomain.server.com": &vhostPair{
				vhostStruct: vhost.New(config.VirtualHost{
					BaseVirtualHost: config.BaseVirtualHost{Name: "subdomain.server.com"},
				}, nil, nil),
			},
			"10.8.3.43": &vhostPair{
				vhostStruct: vhost.New(config.VirtualHost{
					BaseVirtualHost: config.BaseVirtualHost{Name: "10.8.3.43"},
				}, nil, nil),
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

		foundVhost, _ := app.findVirtualHost(req)

		if foundVhost.Name != reqName {
			t.Errorf("Expected to find vhost for %s but it found %s", reqName,
				foundVhost.Name)
		}
	}

	foundVhost, foundHandler := app.findVirtualHost(&http.Request{
		Host: "no-such-host-here.com:993",
	})

	if foundVhost != nil {
		t.Errorf("Searching for non existing virtual host returned one: %s",
			foundVhost.Name)
	}

	if foundHandler != nil {
		t.Error("Searcing for non existing virtual host reutrned a handler")
	}

}
