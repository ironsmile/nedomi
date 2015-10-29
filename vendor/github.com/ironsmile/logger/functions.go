package logger

import (
	"io"
)

// Log calls the default logger's Log function.
// Arguments are handled in the manner of fmt.Print.
func Log(v ...interface{}) {
	defaultLogger.Log(v...)
}

// Logln calls the default logger's Logln function.
// Arguments are handled in the manner of fmt.Println.
func Logln(v ...interface{}) {
	defaultLogger.Logln(v...)
}

// Logf calls the default logger's Logf function.
// Arguments are handled in the manner of fmt.Printf.
func Logf(format string, args ...interface{}) {
	defaultLogger.Logf(format, args...)
}

// Debug calls the default logger's Debug function.
// Arguments are handled in the manner of fmt.Print.
func Debug(v ...interface{}) {
	defaultLogger.Debug(v...)
}

// Debugln calls the default logger's Debugln function.
// Arguments are handled in the manner of fmt.Println.
func Debugln(v ...interface{}) {
	defaultLogger.Debugln(v...)
}

// Debugf calls the default logger's Debugf function.
// Arguments are handled in the manner of fmt.Printf.
func Debugf(format string, args ...interface{}) {
	defaultLogger.Debugf(format, args...)
}

// Error calls the default logger's Error function.
// Arguments are handled in the manner of fmt.Print.
func Error(v ...interface{}) {
	defaultLogger.Error(v...)
}

// Errorln calls the default logger's Errorln function.
// Arguments are handled in the manner of fmt.Println.
func Errorln(v ...interface{}) {
	defaultLogger.Errorln(v...)
}

// Errorf calls the default logger's Errorf function.
// Arguments are handled in the manner of fmt.Printf.
func Errorf(format string, args ...interface{}) {
	defaultLogger.Errorf(format, args...)
}

// Fatal calls the default logger's Fatal function.
// Arguments are handled in the manner of fmt.Print.
func Fatal(v ...interface{}) {
	defaultLogger.Fatal(v...)
}

// Fatalln calls the default logger's Fatalln function.
// Arguments are handled in the manner of fmt.Println.
func Fatalln(v ...interface{}) {
	defaultLogger.Fatalln(v...)
}

// Fatalf calls the default logger's Fatalf function.
// Arguments are handled in the manner of fmt.Printf.
func Fatalf(format string, args ...interface{}) {
	defaultLogger.Fatalf(format, args...)
}

// SetDebugOutput replaces the debug stream output of the default logger.
func SetDebugOutput(w io.Writer) {
	defaultLogger.SetDebugOutput(w)
}

// SetErrorOutput replaces the error stream output of the default logger.
func SetErrorOutput(w io.Writer) {
	defaultLogger.SetErrorOutput(w)
}

// SetLogOutput replaces the log stream output of the default logger.
func SetLogOutput(w io.Writer) {
	defaultLogger.SetLogOutput(w)
}

// SetLevel sets the log level of the default logger.
func SetLevel(level int) {
	defaultLogger.Level = level
}
