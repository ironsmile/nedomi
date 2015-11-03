package netutils

import "net"

type timeoutConnListener struct {
	net.Listener
}

// DeadlineToTimeoutListener wraps the provided listener in a new one
// whose Accept methods returns a wrapped net.Conn whose Deadlines set
// timeouts for each Read|Write individually.
// Example: a conn.SetReadDeadline(time.Now().Add(time.Second)) will set a timeout
// of one second. With the standard conn|listener this will mean that if you start reading a response
// calling Read multiple times but it all takes more than a second it will timeout. With a connection
// from this listener if each call to Read finishes in less than a second the connection will not timeout.
func DeadlineToTimeoutListener(l net.Listener) net.Listener {
	return &timeoutConnListener{Listener: l}
}

// Accept calls the underlying accept and wraps the connection if not nil in timeouting connection
func (t *timeoutConnListener) Accept() (net.Conn, error) {
	conn, err := t.Listener.Accept()
	if conn != nil {
		conn = newTimeoutConn(conn)
	}

	return conn, err
}
