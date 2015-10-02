package std

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/ironsmile/nedomi/config"
)

// New returns a new configured Standard that is ready to use.
func New(cfg *config.Logger) (*Standard, error) {
	if len(cfg.Settings) < 1 {
		return nil, fmt.Errorf("logger 'settings' key is missing")
	}

	var s struct {
		Level string `json:"level"`
	}

	err := json.Unmarshal(cfg.Settings, &s)
	if err != nil {
		return nil, fmt.Errorf("error on parsing settings for 'std' logger:\n%s\n", err)
	}

	var level int
	switch s.Level {
	case "no_log":
		level = NOLOG
	case "info":
		level = INFO
	case "debug":
		level = DEBUG
	case "error":
		level = ERROR
	case "fatal":
		level = FATAL
	default:
		return nil, fmt.Errorf("unsupported log type %s", s.Level)
	}

	return &Standard{
		level: level,
	}, nil
}

// Standard is logging through the standard `log` package
// It can take an additional configuration element 'level'
// with string values of 'no_log', 'fatal', 'error' 'info'
// or 'debug'. The values says which calls to the logger
// to be actually logged. Saying a later value means that
// the previous ones should be logged as well - 'info' means
// that 'fatal' and 'error' should be logged as well.
type Standard struct {
	level int
}

// These are the different logging levels that are supported.
const (
	NOLOG = iota
	FATAL
	ERROR
	INFO
	DEBUG
)

// Log is the same as log.Println if level is atleast 'info'
func (n *Standard) Log(args ...interface{}) {
	if n.level >= INFO {
		log.Println(args...)
	}
}

// Logf is the same as log.Printf, with a newline at the end of format if missing, if level is atleast 'info'
func (n *Standard) Logf(format string, args ...interface{}) {
	if n.level >= INFO {
		log.Println(fmt.Sprintf(format, args...))
	}
}

// Debug is the same as log.Println if level is atleast 'debug'
func (n *Standard) Debug(args ...interface{}) {
	if n.level >= DEBUG {
		log.Println(args...)
	}
}

// Debugf is the same as log.Printf, with a newline at the end of format if missing, if level is atleast 'debug'
func (n *Standard) Debugf(format string, args ...interface{}) {
	if n.level >= DEBUG {
		log.Println(fmt.Sprintf(format, args...))
	}
}

// Error is the same as log.Println if level is atleast 'error'
func (n *Standard) Error(args ...interface{}) {
	if n.level >= ERROR {
		log.Println(args...)
	}
}

// Errorf is the same as log.Printf, with a newline at the end of format if missing, if level is atleast 'error'
func (n *Standard) Errorf(format string, args ...interface{}) {
	if n.level >= ERROR {
		log.Println(fmt.Sprintf(format, args...))
	}
}

// Fatal is the same as log.Fatalln if level is atleast 'fatal'
func (n *Standard) Fatal(args ...interface{}) {
	if n.level >= FATAL {
		log.Fatalln(args...)
	}
}

// Fatalf is the same as log.Fatalf, with a newline at the end of format if missing, if level is atleast 'fatal'
func (n *Standard) Fatalf(format string, args ...interface{}) {
	if n.level >= FATAL {
		log.Fatalln(fmt.Sprintf(format, args...))
	}
}
