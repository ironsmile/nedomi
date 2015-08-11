package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/user"
	"path"
)

// BaseConfig is part of the root configuration type.
type BaseConfig struct {
	System                SystemSection               `json:"system"`
	Logger                LoggerSection               `json:"logger"`
	DefaultCacheType      string                      `json:"default_cache_type"`
	DefaultCacheAlgorithm string                      `json:"default_cache_algorithm"`
	CacheZones            map[string]*json.RawMessage `json:"cache_zones"`
	HTTP                  json.RawMessage             `json:"http"`
}

// Config is the root configuration type. It contains representation for
// everything in config.json.
type Config struct {
	BaseConfig
	CacheZones map[string]*CacheZoneSection `json:"cache_zones"`
	HTTP       *HTTP                        `json:"http"`
}

// UnmarshalJSON is a custom JSON unmashalling that also implements inheritance,
// custom field initiation and data validation for the root config.
func (c *Config) UnmarshalJSON(buff []byte) error {
	if err := json.Unmarshal(buff, &c.BaseConfig); err != nil {
		return err
	}

	// Parse all the cache zones with set default settings
	c.CacheZones = make(map[string]*CacheZoneSection)
	for id, cacheZoneBuff := range c.BaseConfig.CacheZones {
		cacheZone := CacheZoneSection{
			ID:        id,
			Type:      c.DefaultCacheType,
			Algorithm: c.DefaultCacheAlgorithm,
		}

		if err := json.Unmarshal(*cacheZoneBuff, &cacheZone); err != nil {
			return err
		}
		c.CacheZones[id] = &cacheZone
	}

	// Setup the HTTP config
	c.HTTP = &HTTP{parent: c} // Set HTTP's parent to self
	c.HTTP.Logger = c.Logger  // Inherit the logger
	if err := json.Unmarshal(c.BaseConfig.HTTP, &c.HTTP); err != nil {
		return err
	}

	c.BaseConfig.HTTP = nil       // Cleanup
	c.BaseConfig.CacheZones = nil // Cleanup
	return c.Validate()
}

// Validate checks the root config for errors.
func (c *Config) Validate() error {

	if c.System.User != "" {
		if _, err := user.Lookup(c.System.User); err != nil {
			return fmt.Errorf("Invalid `system.user` directive: %s", err)
		}
	}

	if c.System.Pidfile == "" {
		return errors.New("Empty pidfile")
	}

	pidDir := path.Dir(c.System.Pidfile)
	st, err := os.Stat(pidDir)
	if err != nil {
		return fmt.Errorf("Pidfile directory: %s", err)
	}
	if !st.IsDir() {
		return fmt.Errorf("%s is not a directory", pidDir)
	}

	if c.Logger.Type == "" {
		return errors.New("No default logger type found in the `logger` section")
	}

	return nil
}
