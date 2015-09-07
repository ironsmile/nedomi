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
	b.slice = append(b.slice, fmt.Sprint(logPrefix, arg))
}

func (b *Mock) debug(arg string) {
	b.Lock()
	defer b.Unlock()
	b.slice = append(b.slice, fmt.Sprint(debugPrefix, arg))
}
func (b *Mock) err(arg string) {
	b.Lock()
	defer b.Unlock()
	b.slice = append(b.slice, fmt.Sprint(errorPrefix, arg))
}
func (b *Mock) fatal(arg string) {
	b.Lock()
	defer b.Unlock()
	b.slice = append(b.slice, fmt.Sprint(fatalPrefix, arg))
	panic("Fatal called")
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

// Log is the same as log.Println if level is atleast 'info'
func (b *Mock) Log(args ...interface{}) {
	b.log(fmt.Sprint(args...))
}

// Logf is the same as log.Printf, with a '\n' at the end of format if missing, if level is atleast 'info'
func (b *Mock) Logf(format string, args ...interface{}) {
	b.log(fmt.Sprintf(format, args...))
}

// Debug is the same as log.Println if level is atleast 'debug'
func (b *Mock) Debug(args ...interface{}) {
	b.debug(fmt.Sprint(args...))
}

// Debugf is the same as log.Printf, with a '\n' at the end of format if missing, if level is atleast 'debug'
func (b *Mock) Debugf(format string, args ...interface{}) {
	b.debug(fmt.Sprintf(format, args...))
}

// Error is the same as log.Println if level is atleast 'error'
func (b *Mock) Error(args ...interface{}) {
	b.err(fmt.Sprint(args...))
}

// Errorf is the same as log.Printf, with a '\n' at the end of format if missing, if level is atleast 'error'
func (b *Mock) Errorf(format string, args ...interface{}) {
	b.err(fmt.Sprintf(format, args...))
}

// Fatal is the same as log.Fatalln if level is atleast 'fatal'
func (b *Mock) Fatal(args ...interface{}) {
	b.fatal(fmt.Sprint(args...))
}

// Fatalf is the same as log.Fatalf, with a '\n' at the end of format if missing, if level is atleast 'fatal'
func (b *Mock) Fatalf(format string, args ...interface{}) {
	b.fatal(fmt.Sprintf(format, args...))
}

const (
	fatalPrefix = "Fatal:"
	errorPrefix = "Error:"
	logPrefix   = "Log:"
	debugPrefix = "Debug:"
)
