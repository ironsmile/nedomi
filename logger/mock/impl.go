package mock

import (
	"fmt"
	"sync"

	"github.com/ironsmile/nedomi/config"
)

// New returns a new Buffers logger.
func New(cfg *config.LoggerSection) (*Mock, error) {
	b := &Mock{}
	return b, nil
}

// Mock writes all the logs and can return them for later examination
type Mock struct {
	sync.Mutex
	slice []string
}

func (b *Mock) log(arg string) {
	b.Lock()
	defer b.Unlock()
	b.slice = append(b.slice, fmt.Sprint(LogPrefix, arg))
}

func (b *Mock) debug(arg string) {
	b.Lock()
	defer b.Unlock()
	b.slice = append(b.slice, fmt.Sprint(DebugPrefix, arg))
}
func (b *Mock) err(arg string) {
	b.Lock()
	defer b.Unlock()
	b.slice = append(b.slice, fmt.Sprint(ErrorPrefix, arg))
}
func (b *Mock) fatal(arg string) {
	b.Lock()
	defer b.Unlock()
	b.slice = append(b.slice, fmt.Sprint(FatalPrefix, arg))
	panic(FatalPanicText)
}

// Logged returns a slice of strings that is everything that has been logged
// since the creation or the last call to Clear.
func (b *Mock) Logged() []string {
	b.Lock()
	defer b.Unlock()
	result := make([]string, len(b.slice))
	copy(result, b.slice)
	return result
}

// Clear clears the logged messages
func (b *Mock) Clear() {
	b.Lock()
	defer b.Unlock()
	b.slice = make([]string, 0)
}

// Log logs a message with fmt.Sprint with LogPrefix infront
func (b *Mock) Log(args ...interface{}) {
	b.log(fmt.Sprint(args...))
}

// Logf logs a message with fmt.Sprintf with LogPrefix infront
func (b *Mock) Logf(format string, args ...interface{}) {
	b.log(fmt.Sprintf(format, args...))
}

// Debug logs a message with fmt.Sprint adding DebugPrefix infront
func (b *Mock) Debug(args ...interface{}) {
	b.debug(fmt.Sprint(args...))
}

// Debugf logs a message with fmt.Sprintf adding DebugPrefix infront
func (b *Mock) Debugf(format string, args ...interface{}) {
	b.debug(fmt.Sprintf(format, args...))
}

// Error logs a message with fmt.Sprint adding ErrorPrefix infront
func (b *Mock) Error(args ...interface{}) {
	b.err(fmt.Sprint(args...))
}

// Errorf logs a message with fmt.Sprintf adding ErrorPrefix infront
func (b *Mock) Errorf(format string, args ...interface{}) {
	b.err(fmt.Sprintf(format, args...))
}

// Fatal logs a message with fmt.Sprint adding FatalPrefix infront and then panics
func (b *Mock) Fatal(args ...interface{}) {
	b.fatal(fmt.Sprint(args...))
}

// Fatalf logs a message with fmt.Sprintf adding FatalPrefix infront and then panics
func (b *Mock) Fatalf(format string, args ...interface{}) {
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
