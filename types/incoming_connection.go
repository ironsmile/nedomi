package types

// IncomingConn is a type that represents incoming connections
type IncomingConn interface {
	// see net.Conn.RemoteAddr
	ID() string
	// will throttle the connection at the given speed
	SetThrottle(speed BytesSize)
	// stop throttling
	RemoveThrottling()
}
