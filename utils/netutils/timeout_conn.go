package netutils

import (
	"io"
	"net"
	"sync"
	"time"

	"github.com/ironsmile/nedomi/types"
	"github.com/ironsmile/nedomi/utils"
	"github.com/ironsmile/nedomi/utils/throttle"
)

// timeoutConn is a connection for which Deadline sets a timeout equal to
// the difference to which the deadline was set. That timeout is then use to
// Timeout each read|write on the connection
type timeoutConn struct {
	net.Conn
	id                        string
	wr                        io.Writer
	maxSizeOfTransfer         int64
	minSizeOfTransfer         int64
	pool                      sync.Pool
	readTimeout, writeTimeout time.Duration
}

// newTimeoutConn returns a timeout conn wrapping around the provided one
func newTimeoutConn(conn net.Conn, maxSizeOfTransfer, minSizeOfTransfer int64, pool sync.Pool) *timeoutConn {
	return &timeoutConn{
		Conn:              conn,
		maxSizeOfTransfer: maxSizeOfTransfer,
		minSizeOfTransfer: minSizeOfTransfer,
		pool:              pool,
		wr:                conn,
		id:                conn.RemoteAddr().String(),
	}
}

func (tc *timeoutConn) ID() string {
	return tc.id
}

// !TODO conform to maxSizeOfTransfer
func (tc *timeoutConn) Read(data []byte) (int, error) {
	tc.Conn.SetReadDeadline(tc.readDeadline())
	return tc.Conn.Read(data)
}

// !TODO conform to maxSizeOfTransfer
func (tc *timeoutConn) Write(data []byte) (int, error) {
	tc.Conn.SetWriteDeadline(tc.writeDeadline())
	return tc.wr.Write(data)
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

// ReadFrom implementation
func (tc *timeoutConn) ReadFrom(r io.Reader) (n int64, err error) {
	for nn := int64(tc.maxSizeOfTransfer); err == nil && nn == tc.maxSizeOfTransfer; n += nn {
		tc.Conn.SetWriteDeadline(tc.writeDeadline())
		nn, err = utils.CopyN(tc.wr, r, tc.maxSizeOfTransfer)
	}
	if err == io.EOF { // ReadFrom is until EOF
		err = nil
	}
	return
}

func (tc *timeoutConn) writeDeadline() time.Time {
	return time.Now().Add(tc.writeTimeout)
}

func (tc *timeoutConn) readDeadline() time.Time {
	return time.Now().Add(tc.readTimeout)
}

func (tc *timeoutConn) SetThrottle(speed types.BytesSize) {
	tc.wr = throttle.NewThrottleWriter(tc.Conn, int64(speed), tc.minSizeOfTransfer)
}

func (tc *timeoutConn) RemoveThrottling() {
	tc.wr = tc.Conn
}
