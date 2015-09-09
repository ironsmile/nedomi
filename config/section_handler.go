package config

import "encoding/json"

// HandlerSection contains handler options.
type HandlerSection struct {
	Type     string          `json:"type"`
	Settings json.RawMessage `json:"settings"`
}

// Validate checks a HandlerSection config section config for errors.
func (l HandlerSection) Validate() error {
	//!TODO: support flexible type and config check for different modules
	return nil
}

// GetSubsections returns nil (HandlerSection has no subsections).
func (l HandlerSection) GetSubsections() []Section {
	return nil
}
