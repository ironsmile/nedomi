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
	"net/url"
	"os"
	"os/user"
	"path"
	"path/filepath"
	"strconv"
	"strings"
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
	System     SystemSection      `json:"system"`
	Logging    LoggingSection     `json:"logging"`
	HTTP       HTTPSection        `json:"http"`
	CacheZones []CacheZoneSection `json:"cache_zones"`
}

// All configurations conserning the HTTP
type HTTPSection struct {
	Listen  string          `json:"listen"`
	Servers []ServerSection `json:"servers"`
}

type ServerSection struct {
	Name            string `json:"name"`
	UpstreamAddress string `json:"upstream_address"`
	CacheZone       int    `json:"cache_zone"`
	CacheKey        string `json:"cache_key"`
}

type CacheZoneSection struct {
	ID             uint32    `json:"id"`
	Path           string    `json:"path"`
	StorageSize    BytesSize `json:"storage_size"`
	StorageObjects uint64    `json:"storage_objects"`
	PartSize       BytesSize `json:"part_size"`
}

/*
   BytesSize represents size written in string format. Examples: "1m", "20g" etc.
   Its main purpose is to be stored and loaded from json.
*/
type BytesSize uint64

// Bytes returns bytes number as uint64
func (b *BytesSize) Bytes() uint64 {
	return uint64(*b)
}

/*
   Parses bytes size such as "1m", "15g" to BytesSize struct.
*/
func BytesSizeFromString(str string) (BytesSize, error) {

	if len(str) < 1 {
		return 0, errors.New("Size string is too small")
	}

	last := strings.ToLower(str[len(str)-1:])

	sizes := map[string]uint64{
		"":  1,
		"k": 1024,
		"m": 1024 * 1024,
		"g": 1024 * 1024 * 1024,
		"t": 1024 * 1024 * 1024 * 1024,
		"z": 1024 * 1024 * 1024 * 1024 * 1024,
	}

	size, ok := sizes[last]
	var num string

	if ok {
		num = str[:len(str)-1]
	} else {
		num = str
		size = 1
	}

	ret, err := strconv.Atoi(num)

	if err != nil {
		return 0, err
	}

	return BytesSize(uint64(ret) * size), nil
}

func (b *BytesSize) UnmarshalJSON(buff []byte) error {
	var buffStr string
	err := json.Unmarshal(buff, &buffStr)
	if err != nil {
		return err
	}
	parsed, err := BytesSizeFromString(buffStr)
	*b = parsed
	return err
}

// Logging options
type LoggingSection struct {
	LogFile string `json:"log_file"`
	Debug   bool   `json:"debug"`
}

// Contains system and environment configurations.
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

	if cfg.HTTP.Listen == "" {
		return errors.New("Empty listen directive")
	}

	//!TODO: make sure Listen is valid tcp address
	if _, err := net.ResolveTCPAddr("tcp", cfg.HTTP.Listen); err != nil {
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

	for _, serverSection := range cfg.HTTP.Servers {
		parsed, err := url.Parse(serverSection.UpstreamAddress)

		if err != nil {
			return fmt.Errorf("Error parsing server %s upstream. %s",
				serverSection.Name, err)
		}

		if !parsed.IsAbs() {
			return fmt.Errorf("Upstream address was not absolute: %s",
				serverSection.UpstreamAddress)
		}
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
