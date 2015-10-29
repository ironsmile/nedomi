/*
   Package logger delivers a tiered loggin mechanics. It is a thin wrapper above the
   standard library's log package. Its main purpose is to give the developer separate
   log streams and logging severity levels. There are 3 types of streams: debug,
   log and error and 4 levels of logging: LevelDebug, LevelLog, LevelError and
   LevelNoLog.

   Users of the package will interact with its Logger type. It supports the
   Debug(f|ln)?, Log(f|ln)?, Error(f|ln)? and Fatal(f|ln)? functions
   which handle their arguments in the way fmt.Print(f|ln)? do.

   Every Logger has its own log level. See the constants section for information how
   to use this level.

   For fine tuning loggers users will need to import the log package and use the
   Logger.Debugger, Logger.Logger and Logger.Errorer which are just instances
   of log.Logger.

   The package supplies a default logger and a shorthand functions for using it.
   The default logger is created autmatically with its debug, log and error streams
   using stdout, stound and stderr respectively. Its logging level is set to LevelLog.
*/
package logger

import (
	"io"
	"log"
	"os"
)

/*
   These constants are used to configer a Logger's output level. They are ordered in
   a strict ascending order. If the level of a logger is set at particular constant
   it will only emit messages in the streams for this constant's level and above.

   For example if a logger has its Level set to LevelLog it will emit messages
   in its log and error streams but will not emit anything in the debug stream.

   A special level LevelNoLog may be used in order to silence the logger completely.
*/
const (
	LevelDebug = iota // Will show all messages
	LevelLog          // Will show error and log messages
	LevelError        // Will show error messages only
	LevelNoLog        // No logs will be emitted whatsoever
)

var (
	defaultLogger *Logger
)

func init() {
	defaultLogger = New()
}

/*
   Logger is a type which actually consistes of 3 log.Logger object. One for every
   stream - debug, log and error.
*/
type Logger struct {

	/*
	   The debug stream. It is a pointer to the underlying log.Logger strcture
	   which is used when Debug, Debugf and Debugln are called.
	*/
	Debugger *log.Logger

	/*
	   The log stream. Used by Log, Logf and Logln.
	*/
	Logger *log.Logger

	/*
	   The error stream. Used by Error, Errorf and Errorln.
	*/
	Errorer *log.Logger

	/*
	   Which Log Level this Logger is using. See the descrption of the
	   log level constants to understend what this means.
	*/
	Level int
}

/*
   Log emits this message to the log stream.
*/
func (l *Logger) Log(v ...interface{}) {
	if l.Level > LevelLog {
		return
	}
	l.Logger.Print(v...)
}

/*
   Logf emits this message to the log stream. It supports formatting.
   See fmt.Printf for details on the formatting options.
*/
func (l *Logger) Logf(format string, args ...interface{}) {
	if l.Level > LevelLog {
		return
	}
	l.Logger.Printf(format, args...)
}

/*
   Logln emits this message to the log stream and adds a new line at the end of
   the message. Similar to fmt.Println.
*/
func (l *Logger) Logln(v ...interface{}) {
	if l.Level > LevelLog {
		return
	}
	l.Logger.Println(v...)
}

/*
   Debug emits this message to the debug stream.
*/
func (l *Logger) Debug(v ...interface{}) {
	if l.Level > LevelDebug {
		return
	}
	l.Debugger.Print(v...)
}

/*
   Debugf emits this message to the debug stream. Supports fmt.Printf formatting.
*/
func (l *Logger) Debugf(format string, args ...interface{}) {
	if l.Level > LevelDebug {
		return
	}
	l.Debugger.Printf(format, args...)
}

/*
   Debugln emits this message to the debug streamand adds a new line at the end of
   the message. Similar to fmt.Println.
*/
func (l *Logger) Debugln(v ...interface{}) {
	if l.Level > LevelDebug {
		return
	}
	l.Debugger.Println(v...)
}

/*
   Error emits this message to the error stream.
*/
func (l *Logger) Error(v ...interface{}) {
	if l.Level > LevelError {
		return
	}
	l.Errorer.Print(v...)
}

/*
   Errorf emits this message to the error stream. Supports fmt.Printf formatting.
*/
func (l *Logger) Errorf(format string, args ...interface{}) {
	if l.Level > LevelError {
		return
	}
	l.Errorer.Printf(format, args...)
}

/*
   Error emits this message to the error stream and adds a new line at the end of the
   message. Similar to fmt.Println.
*/
func (l *Logger) Errorln(v ...interface{}) {
	if l.Level > LevelError {
		return
	}
	l.Errorer.Println(v...)
}

/*
   Fatal will print the message to the error stream and halt the program.
*/
func (l *Logger) Fatal(v ...interface{}) {
	l.Errorer.Fatal(v...)
}

/*
   Fatalln is similar to Fatal but adds a new line at the end of the output.
*/
func (l *Logger) Fatalln(v ...interface{}) {
	l.Errorer.Fatalln(v...)
}

/*
   Fatalf is similar to Fatal but supports formatting. See fmt.Printf for format
   instructions.
*/
func (l *Logger) Fatalf(format string, v ...interface{}) {
	l.Errorer.Fatalf(format, v...)
}

/*
   SetErrorOutput changes the error output stream. It preserves the flags
   and prefix of the old stream.
*/
func (l *Logger) SetErrorOutput(w io.Writer) {
	l.Errorer = log.New(w, l.Errorer.Prefix(), l.Errorer.Flags())
}

/*
   SetDebugOutput changes the debug output stream. It preserves the flags
   and prefix of the old stream.
*/
func (l *Logger) SetDebugOutput(w io.Writer) {
	l.Debugger = log.New(w, l.Debugger.Prefix(), l.Debugger.Flags())
}

/*
   SetLogOutput changes the log output stream. It preserves the flags
   and prefix of the old stream.
*/
func (l *Logger) SetLogOutput(w io.Writer) {
	l.Logger = log.New(w, l.Logger.Prefix(), l.Logger.Flags())
}

/*
   New creates and returns a new Logger. Its debug and log streams
   are stdout and its error stream is stderr.
   The returned logger will have log level set to LevelLog.
*/
func New() *Logger {
	l := &Logger{}
	l.Debugger = log.New(os.Stdout, "[DEBUG] ", log.LstdFlags)
	l.Logger = log.New(os.Stdout, "[LOG] ", log.LstdFlags)
	l.Errorer = log.New(os.Stderr, "[ERROR] ", log.LstdFlags)
	l.Level = LevelLog
	return l
}

/*
   Default returns a pointer to the default Logger. Using it users can configure
   the behaviour of it.
*/
func Default() *Logger {
	return defaultLogger
}
