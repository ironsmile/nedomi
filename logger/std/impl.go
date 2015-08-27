package std

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/ironsmile/nedomi/config"
)

// New returns configured logger that is ready to use.
func New(cfg *config.LoggerSection) (*stdLogger, error) {
	var s struct {
		Level string `json:"level"`
	}

	err := json.Unmarshal(cfg.Settings, &s)
	if err != nil {
		return nil, fmt.Errorf("Error on parsing settings for 'std' logger:\n%s\n", err)
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
		return nil, fmt.Errorf("Unsupported log error %s", s.Level)
	}

	return &stdLogger{
		level: level,
	}, nil
}

type stdLogger struct {
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

func (n *stdLogger) Log(args ...interface{}) {
	if n.level >= INFO {
		log.Print(args...)
	}
}

func (n *stdLogger) Logf(format string, args ...interface{}) {
	if n.level >= INFO {
		log.Printf(format, args...)
	}
}

func (n *stdLogger) Logln(args ...interface{}) {
	if n.level >= INFO {
		log.Println(args...)
	}
}

func (n *stdLogger) Debug(args ...interface{}) {
	if n.level >= DEBUG {
		log.Print(args...)
	}
}

func (n *stdLogger) Debugf(format string, args ...interface{}) {
	if n.level >= DEBUG {
		log.Printf(format, args...)
	}

}

func (n *stdLogger) Debugln(args ...interface{}) {
	if n.level >= DEBUG {
		log.Println(args...)
	}
}

func (n *stdLogger) Error(args ...interface{}) {
	if n.level >= ERROR {
		log.Print(args...)
	}
}

func (n *stdLogger) Errorf(format string, args ...interface{}) {
	if n.level >= ERROR {
		log.Printf(format, args...)
	}
}

func (n *stdLogger) Errorln(args ...interface{}) {
	if n.level >= ERROR {
		log.Println(args...)
	}
}

func (n *stdLogger) Fatal(args ...interface{}) {
	if n.level >= FATAL {
		log.Fatal(args...)
	}
}

func (n *stdLogger) Fatalf(format string, args ...interface{}) {
	if n.level >= FATAL {
		log.Fatalf(format, args...)
	}
}

func (n *stdLogger) Fatalln(args ...interface{}) {
	if n.level >= FATAL {
		log.Fatalln(args...)
	}
}
