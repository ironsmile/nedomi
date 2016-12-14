package config

import (
	"errors"
	"fmt"
	"os"
	"os/user"
	"path"
)

// System contains system and environment configurations.
type System struct {
	Pidfile string `json:"pidfile"`
	Workdir string `json:"workdir"`
	User    string `json:"user"`
}

// Validate checks a System config section config for errors.
func (s System) Validate() error {
	if s.Pidfile == "" {
		return errors.New("Empty pidfile")
	}

	pidDir := path.Dir(s.Pidfile)
	st, err := os.Stat(pidDir)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("pidfile directory `%s` should be created", pidDir)
		}
		return fmt.Errorf("cannot stat pidfile directory '%s': %s", pidDir, err)
	}
	if !st.IsDir() {
		return fmt.Errorf("%s is not a directory", pidDir)
	}

	if s.User != "" {
		if _, err := user.Lookup(s.User); err != nil {
			return fmt.Errorf("invalid `system.user` directive: %s", err)
		}
	}

	if s.Workdir != "" {
		st, err := os.Stat(s.Workdir)
		if err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("work directory `%s` should be created", s.Workdir)
			}
			return fmt.Errorf("cannot stat work directory '%s': %s", s.Workdir, err)
		}
		if !st.IsDir() {
			return fmt.Errorf("%s is not a directory", s.Workdir)
		}
	}

	return nil
}

// GetSubsections returns nil (System has no subsections).
func (s System) GetSubsections() []Section {
	return nil
}
