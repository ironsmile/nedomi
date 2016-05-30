package ironsmile

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/ironsmile/nedomi/config"

	"github.com/ironsmile/logger"
)

const (
	logFileFlags = os.O_WRONLY | os.O_CREATE | os.O_APPEND
	logFilePerms = 0660
)

// New returns configured ironsmile™ logger that is ready to use.
// Configuration:
// 	error	a path to a file to log calls to Errorf?
// 	log	a path to a file to log calls to Logf?
// 	debug	a path to a file to log calls to Debugf?
//
// If debug is set but log not, debug's file will be used for log.
// If log is set(either through the configuration or copied from  debug),
// but error is not, error will be set to log's file.
// The files are appended to if existing, not truncated.
func New(cfg *config.Logger) (*logger.Logger, error) {
	if len(cfg.Settings) < 1 {
		return nil, fmt.Errorf("logger 'settings' key is missing")
	}

	l := logger.New()
	var s settings
	err := json.Unmarshal(cfg.Settings, &s)
	if err != nil {
		return nil, fmt.Errorf("error while parsing logger settings: %s", err)
	}

	var errorOutput, debugOutput, logOutput *os.File
	var files []*os.File

	if s.DebugFile != "" {
		debugOutput, err = os.OpenFile(s.DebugFile, logFileFlags, logFilePerms)
		if err != nil {
			return nil, fmt.Errorf("error while opening file [%s] for debug output: %s",
				s.DebugFile, err)
		}
		files = append(files, debugOutput)
		l.SetDebugOutput(debugOutput)
	}

	if s.LogFile != "" {
		logOutput, err = os.OpenFile(s.LogFile, logFileFlags, logFilePerms)
		if err != nil {
			return nil, fmt.Errorf("error while opening file [%s] for log output: %s",
				s.LogFile, err)
		}
		files = append(files, logOutput)
	} else if debugOutput != nil {
		logOutput = debugOutput
	}

	if logOutput != nil {
		l.SetLogOutput(logOutput)
	}
	if s.ErrorFile != "" {
		errorOutput, err = os.OpenFile(s.ErrorFile, logFileFlags, logFilePerms)
		if err != nil {
			return nil, fmt.Errorf("Error while opening file [%s] for error output: %s",
				s.ErrorFile, err)
		}
		files = append(files, errorOutput)
	} else if logOutput != nil {
		errorOutput = logOutput
	}
	if errorOutput != nil {
		l.SetErrorOutput(errorOutput)
	}

	if debugOutput != nil {
		l.Level = logger.LevelDebug
	} else if logOutput != nil {
		l.Level = logger.LevelLog
	} else if errorOutput != nil {
		l.Level = logger.LevelError
	} else {
		return nil, fmt.Errorf("ironsmile logger needs at least one file to log to")
	}

	return l, nil
}

type settings struct {
	LogFile   string `json:"log"`
	ErrorFile string `json:"error"`
	DebugFile string `json:"debug"`
}
