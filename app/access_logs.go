package app

import (
	"fmt"
	"io"
	"os"
)

const accessLogFilePerm = 0600

// open an access log with the appropriate permissions on the file
// if it isn't open yet. Return the already open otherwise
func (a accessLogs) openAccessLog(file string) (io.Writer, error) {
	if accessLog, ok := a[file]; ok {
		return accessLog, nil
	}
	accessLog, err := os.OpenFile(
		file,
		os.O_CREATE|os.O_WRONLY|os.O_APPEND,
		accessLogFilePerm,
	)
	if err != nil {
		return nil, fmt.Errorf("error opening access log `%s`- %s",
			file, err)
	}
	a[file] = accessLog
	return accessLog, nil
}

// helper type to facilitate not opening the same access_log twice
type accessLogs map[string]io.Writer
