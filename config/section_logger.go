package config

import (
	"encoding/json"
	"errors"
)

// Logger contains logger options.
type Logger struct {
	Type     string          `json:"type"`
	Settings json.RawMessage `json:"settings"`
}

// Validate checks a Logger config section config for errors.
func (l Logger) Validate() error {

	if l.Type == "" {
		return errors.New("No logger type found in the `logger` section.")
	}

	//!TODO: support flexible type and config check for different modules

	return nil
}

// GetSubsections returns nil (Logger has no subsections).
func (l Logger) GetSubsections() []Section {
	return nil
}
