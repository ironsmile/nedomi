package buffers_test

import (
	"testing"

	"github.com/ironsmile/nedomi/logger/buffers"
)

type commandType int

const (
	Log commandType = iota
	Logf
	Error
	Errorf
	Debug
	Debugf
	Fatal
	Fatalf
	LoggedIs
	Clear
)

type command struct {
	cmd  commandType
	args []interface{}
}

func newCommand(cmd commandType, args ...interface{}) *command {
	return &command{
		cmd:  cmd,
		args: args,
	}
}

//returning false stop execution
func (c *command) testExecute(t *testing.T, b *buffers.Buffers) bool {
	switch c.cmd {
	case Log:
		b.Log(c.args...)
	case Logf:
		b.Logf(c.args[0].(string), c.args[1:]...)
	case Error:
		b.Error(c.args...)
	case Errorf:
		b.Errorf(c.args[0].(string), c.args[1:]...)
	case Debug:
		b.Debug(c.args...)
	case Debugf:
		b.Debugf(c.args[0].(string), c.args[1:]...)
	case Fatal:
		if !catchFatal(func() { b.Fatal(c.args...) }) {
			t.Errorf("fatal didn't panic :(")
		}
	case Fatalf:
		if !catchFatal(func() { b.Fatalf(c.args[0].(string), c.args[1:]...) }) {
			t.Errorf("fatal didn't panic :(")
		}
	case LoggedIs:
		logged := b.Logged()
		expectedLength, actualLength := len(c.args), len(logged)
		if expectedLength != actualLength {
			t.Errorf("Different lengths of expected (%d) and actual (%d) logged messages.", expectedLength, actualLength)
			return false
		}
		for index, expected := range c.args {
			actual := logged[index]
			if expected != actual {
				t.Errorf("Log has wrong contents on index %d expected `%s` but got `%s`", index, expected, actual)
				return false
			}
		}
	case Clear:
		b.Clear()
	default:
		return false
	}
	return true
}

func catchFatal(f func()) (result bool) {
	defer func() {
		if p := recover(); p != "" {
			result = true
		}
	}()
	result = false
	f()
	return
}

func TestLogging(t *testing.T) {
	var testMat = []*command{
		newCommand(LoggedIs),
		newCommand(Log, "arg1"),
		newCommand(Logf, `arg%d%s`, 2, "3"),
		newCommand(LoggedIs, `Log:arg1`, `Log:arg23`),
		newCommand(Error, "err1"),
		newCommand(Debugf, "debug1"),
		newCommand(LoggedIs, `Log:arg1`, `Log:arg23`, `Error:err1`, `Debug:debug1`),
		newCommand(Fatalf, "%s%s%s", "fa", "ta", "l"),
		newCommand(LoggedIs, `Log:arg1`, `Log:arg23`, `Error:err1`, `Debug:debug1`, `Fatal:fatal`),
		newCommand(Clear),
		newCommand(LoggedIs),
	}
	logger, err := buffers.New(nil)
	if err != nil {
		t.Fatalf("Couldn't initialize logger.Buffers - %s", err)
	}

	for index, test := range testMat {
		if !test.testExecute(t, logger) {
			t.Errorf("Command number %d made the test stop", index)
			return
		}
	}
}
