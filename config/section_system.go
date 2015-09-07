package config

import (
	"errors"
	"fmt"
	"os"
	"os/user"
	"path"
)

// SystemSection contains system and environment configurations.
type SystemSection struct {
	Pidfile string `json:"pidfile"`
	Workdir string `json:"workdir"`
	User    string `json:"user"`
}

// Validate checks a SystemSection config section config for errors.
func (s SystemSection) Validate() error {
	if s.Pidfile == "" {
		return errors.New("Empty pidfile")
	}

	pidDir := path.Dir(s.Pidfile)
	st, err := os.Stat(pidDir)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("Pidfile directory `%s` should be created.", pidDir)
		}
		return fmt.Errorf("Cannot stat pidfile directory '%s': %s", pidDir, err)
	}
	if !st.IsDir() {
		return fmt.Errorf("%s is not a directory", pidDir)
	}

	if s.User != "" {
		if _, err := user.Lookup(s.User); err != nil {
			return fmt.Errorf("Invalid `system.user` directive: %s", err)
		}
	}

	if s.Workdir != "" {
		st, err := os.Stat(s.Workdir)
		if err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("Work directory `%s` should be created.", s.Workdir)
			}
			return fmt.Errorf("Cannot stat work directory '%s': %s", s.Workdir, err)
		}
		if !st.IsDir() {
			return fmt.Errorf("%s is not a directory", s.Workdir)
		}
	}

	return nil
}

// GetSubsections returns nil (SystemSection has no subsections).
func (s SystemSection) GetSubsections() []Section {
	return nil
}
