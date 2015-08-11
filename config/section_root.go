package config

import (
	"encoding/json"
	"errors"

	"github.com/davecgh/go-spew/spew"
)

//!TODO: investigate which config options should be pointers and which should be values

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
	return nil
}

// Validate checks the root config for errors.
func (c *Config) Validate() error {

	if len(c.CacheZones) == 0 {
		spew.Dump(c.CacheZones)
		return errors.New("There has to be at least one cache zone")
	}

	return nil
}

// GetSubsections returns a slice with all the subsections of the root config.
func (c *Config) GetSubsections() []Section {
	res := []Section{c.System, c.Logger, c.HTTP}
	for _, cz := range c.CacheZones {
		res = append(res, cz)
	}
	return res
}
