package config

import "encoding/json"

//TODO: split in different files, implement validation?

// LoggerSection contains logger options
type LoggerSection struct {
	Type     string          `json:"type"`
	Settings json.RawMessage `json:"settings"`
}

// SystemSection contains system and environment configurations.
type SystemSection struct {
	Pidfile string `json:"pidfile"`
	Workdir string `json:"workdir"`
	User    string `json:"user"`
}
