# Logger [![GoDoc](https://godoc.org/github.com/ironsmile/logger?status.png)](https://godoc.org/github.com/ironsmile/logger)

This is a simple layer on top of the [standard log package](https://golang.org/pkg/log/). Its main purpose is to give the developer separate log streams and logging severity levels.

The goal of this library is to be as simple as possible to use. It only adds the feature mentioned in the previous paragraph and nothing else.

You can see [this package's documentation at godoc.org](https://godoc.org/github.com/ironsmile/logger). It is much more than what you will find in this README.

## Usage

Logging is as simple as 

```go
import (
    "github.com/ironsmile/logger"
)

func main() {
    logger.Errorf("I have an %d errors", 5)
    logger.SetLevel(logger.LevelDebug)
    logger.Debugln("A debug with a new line")
}
```

The above example uses the default logger. You can create as many loggers as you want:

```go
func main() {
    lgStdOut := logger.New()
    lgStdOut.Log("Log object which logs to the stdout")

    outFile, _ := os.Create("/tmp/file.log")
    lgFile := logger.New()
    lgFile.SetLogOutput(outFile)
    lgFile.Log("This will be written in the log file")

    lgFile.SetErrorOutput(os.Stdout)
    lgFile.Errorln("This will be in the standard output")
}
```

You can take a look at the [example file](logger_test.go) which showcases different usage scenarios.

## Interface

Loggers support the `Debug(f|ln)?`, `Log(f|ln)?`, `Error(f|ln)?` and `Fatal(f|ln)?` functions which handle their arguments in the way `fmt.Print(f|ln)?` do. For more info regarding their arguments see [the documentation](https://godoc.org/github.com/ironsmile/logger).

A Logger is made up of 3 different exported [log.Logger](https://golang.org/pkg/log/#Logger) instances. Every one of them represents a logging stream. The `Logger` structure looks like this:

```go
type Logger struct {
    Debugger *log.Logger
    Logger   *log.Logger
    Errorer  *log.Logger
    Level    int
}
```

As expected `Debugger` represents the debug stream, `Logger` - the logging and `Errorer` - the log stream. Being exported they can be used directly, possibly for configuration.

## Logging Levels

Every Logger has a log level. There are 4 possible levels - LevelDebug, LevelLog, LevelError, LevelNoLog. See the [package documentation](https://godoc.org/github.com/ironsmile/logger#pkg-constants) for information on their usage.

## Configuration

Use [SetDebugOutput](https://godoc.org/github.com/ironsmile/logger#Logger.SetDebugOutput), [SetLogOutput](https://godoc.org/github.com/ironsmile/logger#Logger.SetLogOutput) and [SetErrorOutput](https://godoc.org/github.com/ironsmile/logger#Logger.SetErrorOutput) to change the stream destination.

You can change the logging level by setting the Level property of a logger:
```go
lg := logger.New()
lg.Level = logger.LevelDebug
```

For everything else you can access the underlying [log.Logger](https://golang.org/pkg/log/#Logger) instances directly.

## Fatal functions

They do not have their own output stream but use the error stream. So there is no function `SetFatalOutput` and there is no property `Fataller` in the logger struct.

Underneath they actually use the [log.Fatal](https://golang.org/pkg/log/#Fatal) functions which means a call to `logger.Fatal(f|ln)?` will print the message and halt the program.
