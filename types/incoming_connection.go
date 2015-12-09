package types

import "net"

// IncomingConn is a type that represents incoming connections
type IncomingConn interface {
	// see net.Conn.RemoteAddr
	RemoteAddr() net.Addr
	// will throttle the connection at the given speed
	SetThrottle(speed BytesSize)
	// stop throttling
	RemoveThrottling()
}
