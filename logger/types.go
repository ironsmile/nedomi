package logger

// This file is generated with go generate. Any changes to it will be lost after
// subsequent generates.
//
// If you want to edit it go to types.go.template

import (
	"github.com/ironsmile/nedomi/config"

	"github.com/ironsmile/nedomi/logger/nillogger"
)

type newLoggerFunc func(cfg config.LoggerSection) (Logger, error)

var loggerTypes map[string]newLoggerFunc = map[string]newLoggerFunc{

	"nillogger": func(cfg config.LoggerSection) (Logger, error) {
		return nillogger.New(cfg)
	},
}
