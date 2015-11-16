package netutils

import (
	"io"
	"net"
	"sync"
	"time"
)

// timeoutConn is a connection for which Deadline sets a timeout equal to
// the difference to which the deadline was set. That timeout is then use to
// Timeout each read|write on the connection
type timeoutConn struct {
	net.Conn
	sizeOfTransfer            int64
	byteSlices                sync.Pool
	workers                   *workerPool
	readTimeout, writeTimeout time.Duration
}

// newTimeoutConn returns a timeout conn wrapping around the provided one
func newTimeoutConn(
	conn net.Conn,
	sizeOfTransfer int64,
	byteSlices sync.Pool,
	workers *workerPool) *timeoutConn {
	return &timeoutConn{
		Conn:           conn,
		sizeOfTransfer: sizeOfTransfer,
		byteSlices:     byteSlices,
		workers:        workers,
	}
}

// !TODO conform to maxSizeOfTransfer
func (tc *timeoutConn) Read(data []byte) (n int, err error) {
	var ch = make(chan struct{})
	tc.workers.Exec(funcExecute(func() {
		tc.Conn.SetReadDeadline(time.Now().Add(tc.readTimeout))
		n, err = tc.Conn.Read(data)
		close(ch)
	}))
	<-ch
	return
}

// !TODO conform to maxSizeOfTransfer
func (tc *timeoutConn) Write(data []byte) (n int, err error) {
	var ch = make(chan struct{})
	tc.workers.Exec(funcExecute(func() {
		tc.Conn.SetWriteDeadline(time.Now().Add(tc.writeTimeout))
		n, err = tc.Conn.Write(data)
		close(ch)
	}))
	<-ch
	return
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

// ReadFrom uses the underlying ReadFrom if available or does not use it if not available
func (tc *timeoutConn) ReadFrom(r io.Reader) (n int64, err error) {
	if rf, ok := tc.Conn.(io.ReaderFrom); ok {
		var ch = make(chan struct{})
		var nn int64
		for {
			tc.workers.Exec(funcExecute(func() {
				tc.Conn.SetWriteDeadline(time.Now().Add(tc.writeTimeout))
				nn, err = rf.ReadFrom(io.LimitReader(r, tc.sizeOfTransfer))
				ch <- struct{}{}
			}))
			<-ch
			n += nn
			if err != nil {
				return
			}
			if nn != tc.sizeOfTransfer {
				return
			}
		}
	}

	// this is here because we need to write in smaller pieces in order to set the deadline
	// directly using Copy will not use deadlines if used on the underlying net.Conn
	// or will loop if used on timeoutConn directly
	bufp := tc.byteSlices.Get().(*[]byte)
	var readSize, writeSize int
	var readErr error
	for {
		readSize, readErr = r.Read(*bufp)
		n += int64(readSize)
		writeSize, err = tc.Write((*bufp)[:readSize])
		if err != nil {
			return
		}
		if readSize != writeSize {
			return n, io.ErrShortWrite
		}
		if readErr != nil {
			tc.byteSlices.Put(bufp)
			if readErr == io.EOF {
				return n, nil
			}
			return
		}
	}
}
