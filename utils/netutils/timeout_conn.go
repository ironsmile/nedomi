package netutils

import (
	"net"
	"time"
)

// timeoutConn is a connection for which Deadline sets a timeout equal to
// the difference to which the deadline was set. That timeout is then use to
// Timeout each read|write on the connection
type timeoutConn struct {
	net.Conn
	readTimeout, writeTimeout time.Duration
}

// newTimeoutConn returns a timeout conn wrapping around the provided one
func newTimeoutConn(conn net.Conn) *timeoutConn {
	return &timeoutConn{Conn: conn}
}

func (tc *timeoutConn) Read(data []byte) (int, error) {
	tc.Conn.SetReadDeadline(time.Now().Add(tc.readTimeout))
	return tc.Conn.Read(data)
}

func (tc *timeoutConn) Write(data []byte) (int, error) {
	tc.Conn.SetWriteDeadline(time.Now().Add(tc.writeTimeout))
	return tc.Conn.Write(data)
}

// SetDeadline sets both the read and write timeouts to the difference
// from now to the time provied and calls the underlying SetDeadline
func (tc *timeoutConn) SetDeadline(t time.Time) error {
	tc.readTimeout = t.Sub(time.Now())
	tc.writeTimeout = tc.readTimeout
	return tc.Conn.SetDeadline(t)
}

// SetReadDeadline sets the read timeout to the difference from now
// and the time provided as well as calls the underlying SetReadDeadline
// and returns what it returns
func (tc *timeoutConn) SetReadDeadline(t time.Time) error {
	tc.readTimeout = t.Sub(time.Now())
	return tc.Conn.SetReadDeadline(t)
}

// SetWriteDeadline sets the write timeout to the difference from now
// and the time provided as well as calls the underlying SetWriteDeadline
// and returns what it returns
func (tc *timeoutConn) SetWriteDeadline(t time.Time) error {
	tc.writeTimeout = t.Sub(time.Now())
	return tc.Conn.SetWriteDeadline(t)
}
