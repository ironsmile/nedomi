// Package config is responsible for finding, parsing and verifying the
// application's JSON config.
package config

import (
	"encoding/json"
	"flag"
	"io/ioutil"

	"github.com/ironsmile/nedomi/types"
)

//TODO: create interfaces for configuration sections: json parsing, validation, etc.
// maybe inheritance? these can be used in all modules to validate the config at start

// Path to the configuration file, can be initialized from flags
var configFile types.FilePath

func init() {
	configFile.Set("config.json")
	flag.Var(&configFile, "c", "Configuration file")
}

// parse handles the full parsing and validation of a specified json config file
// and returns a fully populated config struct. The json file is specified by
// the filename argument. Any parsing or validation errors are returned as a
// second parameter.
func parse(filename string) (*Config, error) {
	jsonContents, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	cfg := new(Config)
	err = json.Unmarshal(jsonContents, cfg)
	return cfg, err
}

// Get returns the specified config for the daemon. Any parsing or validation
// errors are returned as a second parameter.
func Get() (*Config, error) {
	return parse(string(configFile))
}
