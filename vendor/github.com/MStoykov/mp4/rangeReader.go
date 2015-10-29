package mp4

import "io"

// RangeReader can be used to get ReadClose for a given range
type RangeReader interface {
	RangeRead(start, length uint64) (io.ReadCloser, error)
}
