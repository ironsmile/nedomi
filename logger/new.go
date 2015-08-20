package logger

// The logTypes map is in types.go and it is generate with `go generate`.
//go:generate go run ../tools/module_generator/main.go -template "types.go.template" -output "types.go"

import (
	"fmt"

	"github.com/ironsmile/nedomi/config"
	"github.com/ironsmile/nedomi/types"
)

// New returns a new Logger ready for use. The lt argument sets which type of
// logger will be returned.
func New(cfg config.LoggerSection) (types.Logger, error) {
	loggerFunc, ok := loggerTypes[cfg.Type]

	if !ok {
		return nil, fmt.Errorf("No such log type: %s", cfg.Type)
	}

	return loggerFunc(cfg)
}
