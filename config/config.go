// Package config is responsible for finding, parsing and verifying the
// application's JSON config.
package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/MStoykov/jsonutils"
	"github.com/ironsmile/nedomi/types"
)

// Path to the configuration file, can be initialized from flags
var configFile types.FilePath

// Section defince the methods each config section has to implement
type Section interface {
	Validate() error
	GetSubsections() []Section
}

func init() {
	if err := configFile.Set("config.json"); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}

	flag.Var(&configFile, "c", "Configuration file")
}

// Parse handles the full parsing and validation of a specified json config file
// and returns a fully populated config struct. The json file is specified by
// the filename argument. Any parsing or validation errors are returned as a
// second parameter.
func Parse(filename string) (*Config, error) {
	jsonContents, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	return parseBytes(jsonContents)
}

func parseBytes(jsonContents []byte) (*Config, error) {
	cfg := new(Config)
	if err := json.Unmarshal(jsonContents, cfg); err != nil {
		return nil, showContextOfError(err, jsonContents)
	}
	return cfg, ValidateRecursive(cfg)
}

// Get returns the specified config for the daemon. Any parsing or validation
// errors are returned as a second parameter.
func Get() (*Config, error) {
	return Parse(string(configFile))
}

// ValidateRecursive validates the supplied configuration section and all of
// its subsections.
func ValidateRecursive(s Section) error {
	if err := s.Validate(); err != nil {
		return err
	}
	for _, subSection := range s.GetSubsections() {
		if err := ValidateRecursive(subSection); err != nil {
			return err
		}
	}
	return nil
}

// Getter is helper type for function which returns a config or
// error if the config supposed to be returned is not parsable or some
// other error was encountered the idea is for this to wrap either
// config.Get or some other way of getting and parsing the config
type Getter func() (*Config, error)

func showContextOfError(err error, jsonContents []byte) error {
	var synErr, ok = err.(*json.SyntaxError)
	if !ok { // nothing to do
		return err
	}

	return jsonutils.NewSyntaxError(synErr, jsonContents, 2)
}
