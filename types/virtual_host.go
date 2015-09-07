package types

// VirtualHost links a config vritual host to its cache algorithm and a storage object.
type VirtualHost struct {
	Location
	Muxer *LocationMuxer
}
