package config

import "encoding/json"

// BaseConfig is part of the root configuration type.
type BaseConfig struct {
	System     SystemSection       `json:"system"`
	Logger     LoggerSection       `json:"logger"`
	HTTP       json.RawMessage     `json:"http"`
	CacheZones []*CacheZoneSection `json:"cache_zones"` //!TODO: why not make this directly into a map? fix config examples also
}

// Config is the root configuration type. It contains representation for
// everything in config.json.
type Config struct {
	BaseConfig
	HTTP HTTPSection `json:"http"`
}

// UnmarshalJSON is a custom JSON unmashalling that also implements inheritance,
// custom field initiation and data validation for the root config.
func (c *Config) UnmarshalJSON(buff []byte) error {
	if err := json.Unmarshal(buff, &c.BaseConfig); err != nil {
		return err
	}
	c.HTTP = HTTPSection{parent: c} // Set HTTP's parent to self
	c.HTTP.Logger = c.Logger        // Inherit the logger

	if err := json.Unmarshal(c.BaseConfig.HTTP, &c.HTTP); err != nil {
		return err
	}

	c.BaseConfig.HTTP = nil // Cleanup
	return c.Validate()
}

// Validate checks the root config for logical errors.
func (c *Config) Validate() error {
	//!TODO: implement
	return nil
}
