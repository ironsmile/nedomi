package httputils

import (
	"errors"
	"strconv"
	"strings"
)

// ContentRange specifies the byte range to be sent to the client.
type ContentRange struct {
	Start, Length, ObjSize uint64
}

// ParseResponseContentRange parses a "Content-Range" header string. It only
// implements a subset of RFC 7233 - asterisks (unknown complete-length or
// unsatisfied-range) are treated as an error.
func ParseResponseContentRange(cr string) (*ContentRange, error) {
	if cr == "" {
		return nil, nil // header not present
	}
	const b = "bytes "
	if !strings.HasPrefix(cr, b) {
		return nil, errors.New("invalid range")
	}
	cr = strings.TrimSpace(cr[len(b):])

	i := strings.Index(cr, "/")
	if i < 0 {
		return nil, errors.New("invalid range")
	}
	size, err := strconv.ParseUint(strings.TrimSpace(cr[i+1:]), 10, 64)
	if err != nil {
		return nil, err
	}
	cr = strings.TrimSpace(cr[:i])
	i = strings.Index(cr, "-")
	if i < 0 {
		return nil, errors.New("invalid range")
	}

	start, err := strconv.ParseUint(strings.TrimSpace(cr[:i]), 10, 64)
	if err != nil {
		return nil, err
	}
	end, err := strconv.ParseUint(strings.TrimSpace(cr[i+1:]), 10, 64)
	if err != nil {
		return nil, err
	}

	if start > end || end >= size {
		return nil, errors.New("invalid range")
	}
	return &ContentRange{start, end - start + 1, size}, nil
}
