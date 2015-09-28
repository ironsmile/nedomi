package config

import (
	"encoding/json"
	"errors"
)

type loggerBase struct {
	Type     string          `json:"type"`
	Settings json.RawMessage `json:"settings"`
}

// Logger contains logger options.
type Logger struct {
	loggerBase
}

// UnmarshalJSON is a custom JSON unmarshalling where custom stands for resetting the Settings field.
func (l *Logger) UnmarshalJSON(buff []byte) error {
	l.Settings = append(json.RawMessage{}, l.Settings...)
	return json.Unmarshal(buff, &l.loggerBase)
}

// NewLogger creates a new Logger with the provided type name and settings
func NewLogger(name string, settings json.RawMessage) *Logger {
	return &Logger{
		loggerBase: loggerBase{
			Type:     name,
			Settings: settings,
		},
	}
}

// Validate checks a Logger config section config for errors.
func (l Logger) Validate() error {
	if l.Type == "" {
		return errors.New("No logger type found in the `logger` section.")
	}

	return nil
}

// GetSubsections returns nil (Logger has no subsections).
func (l Logger) GetSubsections() []Section {
	return nil
}
