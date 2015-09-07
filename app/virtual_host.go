package app

import "github.com/ironsmile/nedomi/types"

// VirtualHost links a config vritual host to its cache algorithm and a storage object.
type VirtualHost struct {
	types.Location
	Muxer *LocationMuxer
}
