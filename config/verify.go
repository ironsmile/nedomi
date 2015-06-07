package config

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"os"
	"os/user"
	"path"
)

// Checks all fields in the parsed configs for wrong values. If found, returns error
// explaining the problem.
func (cfg *Config) Verify() error {

	if cfg.System.User != "" {
		if _, err := user.Lookup(cfg.System.User); err != nil {
			return fmt.Errorf("Wrong system.user directive. %s", err)
		}
	}

	if cfg.HTTP.Listen == "" {
		return errors.New("Empty listen directive")
	}

	//!TODO: make sure Listen is valid tcp address
	if _, err := net.ResolveTCPAddr("tcp", cfg.HTTP.Listen); err != nil {
		return err
	}

	if cfg.System.Pidfile == "" {
		return errors.New("Empty pidfile")
	}

	pidDir := path.Dir(cfg.System.Pidfile)
	st, err := os.Stat(pidDir)

	if err != nil {
		return fmt.Errorf("Pidfile directory: %s", err)
	}

	if !st.IsDir() {
		return fmt.Errorf("%s is not a directory", pidDir)
	}

	cacheZonesMap := make(map[uint32]*CacheZoneSection)

	for _, cz := range cfg.CacheZones {
		cacheZonesMap[cz.ID] = cz
	}

	for _, virtualHost := range cfg.HTTP.Servers {

		if !virtualHost.IsForProxyModule() {
			continue
		}

		parsed, err := url.Parse(virtualHost.UpstreamAddress)

		if err != nil {
			return fmt.Errorf("Error parsing server %s upstream. %s",
				virtualHost.Name, err)
		}

		if !parsed.IsAbs() {
			return fmt.Errorf("Upstream address was not absolute: %s",
				virtualHost.UpstreamAddress)
		}

		virtualHost.upstreamAddressUrl = parsed

		if cz, ok := cacheZonesMap[virtualHost.CacheZone]; ok {
			virtualHost.cacheZone = cz
		} else {
			return fmt.Errorf("Upstream %s has not existing cache zone id. %d",
				virtualHost.Name, virtualHost.CacheZone)
		}
	}

	return nil
}
