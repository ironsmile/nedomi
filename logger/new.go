package logger

// The logTypes map is in types.go and it is generate with `go generate`.
//go:generate go run ../tools/module_generator/main.go -template "types.go.template" -output "types.go"

import (
	"fmt"

	"github.com/ironsmile/nedomi/config"
)

/*
   New returns a new Logger ready for use. The lt argument sets which type of logger
   will be returned.
*/
func New(lt string, cfg config.LoggerSection) (Logger, error) {
	loggerFunc, ok := loggerTypes[lt]

	if !ok {
		return nil, fmt.Errorf("No such log type: %s", lt)
	}

	return loggerFunc(cfg)
}
