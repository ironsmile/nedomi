package config

import (
	"encoding/json"
	"errors"
)

// LoggerSection contains logger options.
type LoggerSection struct {
	Type     string          `json:"type"`
	Settings json.RawMessage `json:"settings"`
}

// Validate checks a LoggerSection config section config for errors.
func (l LoggerSection) Validate() error {

	if l.Type == "" {
		return errors.New("No logger type found in the `logger` section.")
	}

	//!TODO: support flexible type and config check for different modules

	return nil
}

// GetSubsections returns nil (LoggerSection has no subsections).
func (l LoggerSection) GetSubsections() []Section {
	return nil
}
