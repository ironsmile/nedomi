/*
   Package config is responsible for finding, parsing and verifying the
   application's JSON config.
*/
package config

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/user"
	"path"
	"path/filepath"
)

// Path to the configuration file, initialized from flags
var ConfigFile string

func init() {
	binaryAbsolute, err := filepath.Abs(os.Args[0])
	if err != nil {
		log.Fatalln("Was not able to find absolute path to config")
	}
	defaultConfigPath := filepath.Join(filepath.Dir(binaryAbsolute), "config.json")
	flag.StringVar(&ConfigFile, "c", defaultConfigPath, "Configuration file")
}

// the configuration type. Should contain representation for everything in config.json
type Config struct {
	Listen  string         `json:"listen"`
	System  SystemSection  `json:"system"`
	Logging LoggingSection `json:"logging"`
}

// Logging options
type LoggingSection struct {
	LogFile string `json:"log_file"`
	Debug   bool   `json:"debug"`
}

type SystemSection struct {
	Pidfile string `json:"pidfile"`
	Workdir string `json:"workdir"`
	User    string `json:"user"`
}

// The config object parses an json file and populates its fields.
// The json file is specified by the filename argument.
func (cfg *Config) Parse(filename string) error {
	jsonContents, err := ioutil.ReadFile(filename)

	if err != nil {
		return err
	}

	return json.Unmarshal(jsonContents, cfg)
}

// Checks all fields in the parsed configs for wrong values. If found, returns error
// explaining the problem.
func (cfg *Config) Verify() error {

	if cfg.System.User != "" {
		if _, err := user.Lookup(cfg.System.User); err != nil {
			return fmt.Errorf("Wrong system.user directive. %s", err)
		}
	}

	if cfg.Listen == "" {
		return errors.New("Empty listen directive")
	}

	//!TODO: make sure Listen is valid tcp address
	if _, err := net.ResolveTCPAddr("tcp", cfg.Listen); err != nil {
		return err
	}

	if cfg.System.Pidfile == "" {
		return errors.New("Empty pidfile")
	}

	pidDir := path.Dir(cfg.System.Pidfile)
	st, err := os.Stat(pidDir)

	if err != nil {
		return fmt.Errorf("Pidfile directory: %s", err)
	}

	if !st.IsDir() {
		return fmt.Errorf("%s is not a directory", pidDir)
	}

	return nil
}

// Finds and returns the config for the daemon. Any errors are returned as a second
// parameter.
func Get() (*Config, error) {
	cfg := &Config{}
	err := cfg.Parse(ConfigFile)
	if err != nil {
		return nil, err
	}
	return cfg, cfg.Verify()
}
