/*
   Package config is responsible for finding, parsing and verifying the
   application's JSON config.
*/
package config

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"net/url"
	"os"
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
	System     SystemSection       `json:"system"`
	Logging    LoggingSection      `json:"logging"`
	HTTP       HTTPSection         `json:"http"`
	CacheZones []*CacheZoneSection `json:"cache_zones"`
}

// All configurations conserning the HTTP
type HTTPSection struct {
	Listen         string         `json:"listen"`
	Servers        []*VirtualHost `json:"virtual_hosts"`
	MaxHeadersSize int            `json:"max_headers_size"`
	ReadTimeout    uint32         `json:"read_timeout"`
	WriteTimeout   uint32         `json:"write_timeout"`
	StatusPage     string         `json:"status_page"`
	CacheAlgo      string         `json:"cache_algorithm"`
	UpstreamType   string         `json:"upstream_type"`
}

type VirtualHost struct {
	Name            string `json:"name"`
	UpstreamAddress string `json:"upstream_address"`
	CacheZone       uint32 `json:"cache_zone"`
	CacheKey        string `json:"cache_key"`
	HandlerType     string `json:"handler"`
	UpstreamType    string `json:"upstream_type"`

	// used internally
	upstreamAddressUrl *url.URL
	cacheZone          *CacheZoneSection
}

type CacheZoneSection struct {
	ID             uint32    `json:"id"`
	Path           string    `json:"path"`
	StorageObjects uint64    `json:"storage_objects"`
	PartSize       BytesSize `json:"part_size"`
	CacheAlgo      string    `json:"cache_algorithm"`
}

// UpstreamUrl returns the previously calculated *url.URL of the upstream
// attached to this VirtualHost.
func (vh *VirtualHost) UpstreamUrl() *url.URL {
	return vh.upstreamAddressUrl
}

// GetCacheZoneSection returns config.CacheZoneSection for this virtual host.
func (vh *VirtualHost) GetCacheZoneSection() *CacheZoneSection {
	return vh.cacheZone
}

// IsForProxyModule returns true if the virtual host should use the default
// proxy handler module as its handler. False otherwise.
func (vh *VirtualHost) IsForProxyModule() bool {
	return vh.HandlerType == "" || vh.HandlerType == "proxy"
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
