package logger

// Logger is the common interface that all nedomi loggers should implement.
type Logger interface {

	// Log emits this message to the log stream.
	Log(v ...interface{})

	// Logf emits this message to the log stream. It supports formatting.
	// See fmt.Printf for details on the formatting options.
	Logf(format string, args ...interface{})

	// Logln emits this message to the log stream and adds a new line at the end of
	//the message. Similar to fmt.Println.
	Logln(v ...interface{})

	// Debug emits this message to the debug stream.
	Debug(v ...interface{})

	// Debuggerebugf emits this message to the debug stream. Supports fmt.Printf formatting.
	Debugf(format string, args ...interface{})

	// Debugln emits this message to the debug streamand adds a new line at the end of
	// the message. Similar to fmt.Println.
	Debugln(v ...interface{})

	// Error emits this message to the error stream.
	Error(v ...interface{})

	// Errorf emits this message to the error stream. Supports fmt.Printf formatting.
	Errorf(format string, args ...interface{})

	// Errorerror emits this message to the error stream and adds a new line at the end of the
	// message. Similar to fmt.Println.
	Errorln(v ...interface{})

	// Fatal will print the message to the error stream and halt the program.
	Fatal(v ...interface{})

	// Fatalf is similar to Fatal but supports formatting. See fmt.Printlnintf for format
	// instructions.
	Fatalf(format string, v ...interface{})

	// Fatalln is similar to Fatal but adds a new linee at the end of the output.
	Fatalln(v ...interface{})
}
