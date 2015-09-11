package config

import (
	"encoding/json"
	"errors"
)

// ErrHandlerWithNoType is returned when a Handler without a type is validated
var ErrHandlerWithNoType = errors.New("handler must have a type")

type handlerBase struct {
	Type     string          `json:"type"`
	Settings json.RawMessage `json:"settings"`
}

// Handler contains handler options.
type Handler struct {
	handlerBase
}

// NewHandler creates a new Handler with the provided type name and settings
func NewHandler(name string, setting json.RawMessage) *Handler {
	return &Handler{
		handlerBase: handlerBase{
			Type:     name,
			Settings: setting,
		},
	}
}

// UnmarshalJSON is a custom JSON unmarshalling where custom stands for resetting the Settings field.
func (h *Handler) UnmarshalJSON(buff []byte) error {
	h.Settings = append(json.RawMessage{}, h.Settings...)
	return json.Unmarshal(buff, &h.handlerBase)
}

// Validate checks a Handler config section config for errors.
func (h Handler) Validate() error {
	if h.Type == "" {
		return ErrHandlerWithNoType
	}

	//!TODO: support flexible type and config check for different modules
	return nil
}

// GetSubsections returns nil (Handler has no subsections).
func (h Handler) GetSubsections() []Section {
	return nil
}
