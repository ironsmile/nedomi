package mock

import (
	"fmt"
	"sync"
)

// NewLogger returns a new Buffers logger.
func NewLogger() *Logger {
	return &Logger{}
}

// Logger writes all the logs and can return them for later examination
type Logger struct {
	sync.Mutex
	slice []string
}

func (b *Logger) log(arg string) {
	b.Lock()
	defer b.Unlock()
	b.slice = append(b.slice, fmt.Sprint(LogPrefix, arg))
}

func (b *Logger) debug(arg string) {
	b.Lock()
	defer b.Unlock()
	b.slice = append(b.slice, fmt.Sprint(DebugPrefix, arg))
}
func (b *Logger) err(arg string) {
	b.Lock()
	defer b.Unlock()
	b.slice = append(b.slice, fmt.Sprint(ErrorPrefix, arg))
}
func (b *Logger) fatal(arg string) {
	b.Lock()
	defer b.Unlock()
	b.slice = append(b.slice, fmt.Sprint(FatalPrefix, arg))
	panic(FatalPanicText)
}

// Logged returns a slice of strings that is everything that has been logged
// since the creation or the last call to Clear.
func (b *Logger) Logged() []string {
	b.Lock()
	defer b.Unlock()
	result := make([]string, len(b.slice))
	copy(result, b.slice)
	return result
}

// Clear clears the logged messages
func (b *Logger) Clear() {
	b.Lock()
	defer b.Unlock()
	b.slice = make([]string, 0)
}

// Log logs a message with fmt.Sprint with LogPrefix infront
func (b *Logger) Log(args ...interface{}) {
	b.log(fmt.Sprint(args...))
}

// Logf logs a message with fmt.Sprintf with LogPrefix infront
func (b *Logger) Logf(format string, args ...interface{}) {
	b.log(fmt.Sprintf(format, args...))
}

// Debug logs a message with fmt.Sprint adding DebugPrefix infront
func (b *Logger) Debug(args ...interface{}) {
	b.debug(fmt.Sprint(args...))
}

// Debugf logs a message with fmt.Sprintf adding DebugPrefix infront
func (b *Logger) Debugf(format string, args ...interface{}) {
	b.debug(fmt.Sprintf(format, args...))
}

// Error logs a message with fmt.Sprint adding ErrorPrefix infront
func (b *Logger) Error(args ...interface{}) {
	b.err(fmt.Sprint(args...))
}

// Errorf logs a message with fmt.Sprintf adding ErrorPrefix infront
func (b *Logger) Errorf(format string, args ...interface{}) {
	b.err(fmt.Sprintf(format, args...))
}

// Fatal logs a message with fmt.Sprint adding FatalPrefix infront and then panics
func (b *Logger) Fatal(args ...interface{}) {
	b.fatal(fmt.Sprint(args...))
}

// Fatalf logs a message with fmt.Sprintf adding FatalPrefix infront and then panics
func (b *Logger) Fatalf(format string, args ...interface{}) {
	b.fatal(fmt.Sprintf(format, args...))
}

// Prefixes for the logged messages
const (
	FatalPrefix = "Fatal:"
	ErrorPrefix = "Error:"
	LogPrefix   = "Log:"
	DebugPrefix = "Debug:"
)

// FatalPanicText is the text of the panic when Fatalf? is called
const FatalPanicText = "Fatal called"
