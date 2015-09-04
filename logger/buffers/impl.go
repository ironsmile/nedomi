package buffers

import (
	"bytes"
	"fmt"

	"github.com/ironsmile/nedomi/config"
)

// New returns a new Buffers logger.
func New(cfg *config.LoggerSection) (*Buffers, error) {
	b := &Buffers{
		buffer: new(bytes.Buffer),
	}
	return b, nil
}

// Buffers writes all the logs to
// bytes.Buffer for later examination
type Buffers struct {
	buffer *bytes.Buffer
}

func (b *Buffers) log(arg string) {
	fmt.Fprintln(b.buffer, logPrefix, arg)
}

func (b *Buffers) debug(arg string) {
	fmt.Fprintln(b.buffer, debugPrefix, arg)
}
func (b *Buffers) err(arg string) {
	fmt.Fprintln(b.buffer, errorPrefix, arg)
}
func (b *Buffers) fatal(arg string) {
	fmt.Fprintln(b.buffer, fatalPrefix, arg)
	panic("Fatal called")
}

// Log is the same as log.Println if level is atleast 'info'
func (b *Buffers) Log(args ...interface{}) {
	b.log(fmt.Sprint(args...))
}

// Logf is the same as log.Printf, with a '\n' at the end of format if missing, if level is atleast 'info'
func (b *Buffers) Logf(format string, args ...interface{}) {
	b.log(fmt.Sprintf(format, args...))
}

// Debug is the same as log.Println if level is atleast 'debug'
func (b *Buffers) Debug(args ...interface{}) {
	b.debug(fmt.Sprint(args...))
}

// Debugf is the same as log.Printf, with a '\n' at the end of format if missing, if level is atleast 'debug'
func (b *Buffers) Debugf(format string, args ...interface{}) {
	b.debug(fmt.Sprintf(format, args...))
}

// Error is the same as log.Println if level is atleast 'error'
func (b *Buffers) Error(args ...interface{}) {
	b.err(fmt.Sprint(args...))
}

// Errorf is the same as log.Printf, with a '\n' at the end of format if missing, if level is atleast 'error'
func (b *Buffers) Errorf(format string, args ...interface{}) {
	b.err(fmt.Sprintf(format, args...))
}

// Fatal is the same as log.Fatalln if level is atleast 'fatal'
func (b *Buffers) Fatal(args ...interface{}) {
	b.fatal(fmt.Sprint(args...))
}

// Fatalf is the same as log.Fatalf, with a '\n' at the end of format if missing, if level is atleast 'fatal'
func (b *Buffers) Fatalf(format string, args ...interface{}) {
	b.fatal(fmt.Sprintf(format, args...))
}

const (
	fatalPrefix = "Fatal:"
	errorPrefix = "Error:"
	logPrefix   = "Log:"
	debugPrefix = "Debug:"
)
