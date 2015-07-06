package std

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/ironsmile/nedomi/config"
)

func New(cfg config.LoggerSection) (*stdLogger, error) {
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
		level = NO_LOG
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

const (
	NO_LOG = iota
	FATAL
	ERROR
	INFO
	DEBUG
)

func (n *stdLogger) Log(v ...interface{}) {
	if n.level >= INFO {
		log.Print(v)
	}
}

func (n *stdLogger) Logf(format string, args ...interface{}) {
	if n.level >= INFO {
		log.Printf(format, args)
	}
}

func (n *stdLogger) Logln(v ...interface{}) {
	if n.level >= INFO {
		log.Println(v)
	}
}

func (n *stdLogger) Debug(v ...interface{}) {
	if n.level >= DEBUG {
		log.Print(v)
	}
}

func (n *stdLogger) Debugf(format string, args ...interface{}) {
	if n.level >= DEBUG {
		log.Printf(format, args)
	}

}

func (n *stdLogger) Debugln(v ...interface{}) {
	if n.level >= DEBUG {
		log.Println(v)
	}
}

func (n *stdLogger) Error(v ...interface{}) {
	if n.level >= ERROR {
		log.Print(v)
	}
}

func (n *stdLogger) Errorf(format string, args ...interface{}) {
	if n.level >= ERROR {
		log.Printf(format, args)
	}
}

func (n *stdLogger) Errorln(v ...interface{}) {
	if n.level >= ERROR {
		log.Println(v)
	}
}

func (n *stdLogger) Fatal(v ...interface{}) {
	if n.level >= FATAL {
		log.Fatal(v)
	}
}

func (n *stdLogger) Fatalf(format string, v ...interface{}) {
	if n.level >= FATAL {
		log.Fatalf(format, v)
	}
}

func (n *stdLogger) Fatalln(v ...interface{}) {
	if n.level >= FATAL {
		log.Fatalln(v)
	}
}
