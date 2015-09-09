package config

import (
	"encoding/json"
	"errors"
)

// ErrHandlerWithNoType is returned when a Handler without a type is validated
var ErrHandlerWithNoType = errors.New("handler must have a type")

// Handler contains handler options.
type Handler struct {
	Type     string          `json:"type"`
	Settings json.RawMessage `json:"settings"`
}

// Validate checks a Handler config section config for errors.
func (l Handler) Validate() error {
	if l.Type == "" {
		return ErrHandlerWithNoType
	}

	//!TODO: support flexible type and config check for different modules
	return nil
}

// GetSubsections returns nil (Handler has no subsections).
func (l Handler) GetSubsections() []Section {
	return nil
}
