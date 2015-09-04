package buffers

import (
	"testing"
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
func (c *command) testExecute(t *testing.T, buffers *Buffers) bool {
	switch c.cmd {
	case Log:
		buffers.Log(c.args...)
	case Logf:
		buffers.Logf(c.args[0].(string), c.args[1:]...)
	case Error:
		buffers.Error(c.args...)
	case Errorf:
		buffers.Errorf(c.args[0].(string), c.args[1:]...)
	case Debug:
		buffers.Debug(c.args...)
	case Debugf:
		buffers.Debugf(c.args[0].(string), c.args[1:]...)
	case Fatal:
		if !catchFatal(func() { buffers.Fatal(c.args...) }) {
			t.Errorf("fatal didn't panic :(")
		}
	case Fatalf:
		if !catchFatal(func() { buffers.Fatalf(c.args[0].(string), c.args[1:]...) }) {
			t.Errorf("fatal didn't panic :(")
		}
	case LoggedIs:
		got := buffers.buffer.String()
		expected := c.args[0].(string)
		if got != expected {
			t.Errorf("LogBuffer has wrong contents!\nexpected : ```\n%s```\ngot : ```\n%s```", got, expected)
			return false
		}
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
		newCommand(Log, "arg1"),
		newCommand(Logf, `arg%d%s`, 2, "3"),
		newCommand(LoggedIs, `Log: arg1
Log: arg23
`),
		newCommand(Error, "err1"),
		newCommand(Debugf, "debug1"),
		newCommand(LoggedIs, `Log: arg1
Log: arg23
Error: err1
Debug: debug1
`),
		newCommand(Fatalf, "%s%s%s", "fa", "ta", "l"),
		newCommand(LoggedIs, `Log: arg1
Log: arg23
Error: err1
Debug: debug1
Fatal: fatal
`),
	}
	logger, err := New(nil)
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
