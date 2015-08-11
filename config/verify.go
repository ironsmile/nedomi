package config

import (
	"errors"
	"fmt"
	"net"
	"os"
	"os/user"
	"path"
)

// Verify checks all fields in the parsed configs for wrong values. If found,
// it returns an error explaining the problem.
func (cfg *Config) Verify() error {

	//!TODO: move to the different sections' Validate() functions

	if cfg.System.User != "" {
		if _, err := user.Lookup(cfg.System.User); err != nil {
			return fmt.Errorf("Wrong system.user directive. %s", err)
		}
	}

	if cfg.Logger.Type == "" {
		return errors.New("No default logger type found in the `logger` section")
	}

	if cfg.HTTP.Listen == "" {
		return errors.New("Empty listen directive")
	}

	if cfg.HTTP.DefaultUpstreamType == "" {
		return errors.New("No default upstream type found in the `http` section")
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

	return nil
}
