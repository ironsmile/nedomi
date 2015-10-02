package ironsmile

import (
	"encoding/json"
	"fmt"
	"io"
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

	logger := logger.New()
	var s settings
	err := json.Unmarshal(cfg.Settings, &s)
	if err != nil {
		return nil, fmt.Errorf("error while parsing logger settings: %s", err)
	}

	var errorOutput, debugOutput, logOutput io.Writer

	if s.DebugFile != "" {
		debugOutput, err = os.OpenFile(s.DebugFile, logFileFlags, logFilePerms)
		if err != nil {
			return nil, fmt.Errorf("error while opening file [%s] for debug output: %s",
				s.DebugFile, err)
		}
		logger.SetDebugOutput(debugOutput)
	}

	if s.LogFile != "" {
		logOutput, err = os.OpenFile(s.LogFile, logFileFlags, logFilePerms)
		if err != nil {
			return nil, fmt.Errorf("error while opening file [%s] for log output: %s",
				s.LogFile, err)
		}
		logger.SetLogOutput(logOutput)
	} else if debugOutput != nil {
		logger.SetLogOutput(debugOutput)
	}

	if s.ErrorFile != "" {
		errorOutput, err = os.OpenFile(s.ErrorFile, logFileFlags, logFilePerms)
		if err != nil {
			return nil, fmt.Errorf("Error while opening file [%s] for error output: %s",
				s.ErrorFile, err)
		}
		logger.SetErrorOutput(errorOutput)
	} else if logOutput != nil {
		logger.SetErrorOutput(logOutput)
	}

	return logger, nil
}

type settings struct {
	LogFile   string `json:"log"`
	ErrorFile string `json:"error"`
	DebugFile string `json:"debug"`
}
