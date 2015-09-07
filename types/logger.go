package types

// Logger is the common interface that all nedomi loggers should implement.
type Logger interface {
	// Log emits this message to the log stream.
	Log(v ...interface{})

	// Logf emits this message to the log stream. Supports fmt.Printf formating.
	Logf(format string, args ...interface{})

	// Debug emits this message to the debug stream.
	Debug(v ...interface{})

	// Debuggerebugf emits this message to the debug stream. Supports fmt.Printf formatting.
	Debugf(format string, args ...interface{})

	// Error emits this message to the error stream.
	Error(v ...interface{})

	// Errorf emits this message to the error stream. Supports fmt.Printf formatting.
	Errorf(format string, args ...interface{})

	// Fatal will print the message to the error stream and halt the program.
	Fatal(v ...interface{})

	// Fatalf is similar to Fatal but supports formatting. Support fmt.Printf formating.
	Fatalf(format string, v ...interface{})
}
