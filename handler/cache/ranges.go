package cache

// This file has been based on http://golang.org/src/net/http/fs.go

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// httpRange specifies the byte range to be sent to the client.
type httpRange struct {
	start, length uint64
}

func (r httpRange) contentRange(size uint64) string {
	return fmt.Sprintf("bytes %d-%d/%d", r.start, r.start+r.length-1, size)
}

// parseReqRange parses a client "Range" header string as per RFC 7233.
func parseReqRange(s string, size uint64) ([]httpRange, error) {
	if s == "" {
		return nil, nil // header not present
	}
	const b = "bytes="
	if !strings.HasPrefix(s, b) {
		return nil, errors.New("invalid range")
	}
	var ranges []httpRange
	for _, ra := range strings.Split(s[len(b):], ",") {
		ra = strings.TrimSpace(ra)
		if ra == "" {
			continue
		}
		i := strings.Index(ra, "-")
		if i < 0 {
			return nil, errors.New("invalid range")
		}
		start, end := strings.TrimSpace(ra[:i]), strings.TrimSpace(ra[i+1:])
		var r httpRange
		if start == "" {
			// If no start is specified, end specifies the
			// range start relative to the end of the file.
			i, err := strconv.ParseUint(end, 10, 64)
			if err != nil {
				return nil, errors.New("invalid range")
			}
			if i > size {
				i = size
			}
			r.start = size - i
			r.length = size - r.start
		} else {
			i, err := strconv.ParseUint(start, 10, 64)
			if err != nil || i >= size || i < 0 {
				return nil, errors.New("invalid range")
			}
			r.start = i
			if end == "" {
				// If no end is specified, range extends to end of the file.
				r.length = size - r.start
			} else {
				i, err := strconv.ParseUint(end, 10, 64)
				if err != nil || r.start > i {
					return nil, errors.New("invalid range")
				}
				if i >= size {
					i = size - 1
				}
				r.length = i - r.start + 1
			}
		}
		ranges = append(ranges, r)
	}
	return ranges, nil
}

// httpContentRange specifies the byte range to be sent to the client.
type httpContentRange struct {
	start, end, size uint64
}

// parseRespContentRange parses a "Content-Range" header string as per RFC 7233.
func parseRespContentRange(cr string) (*httpContentRange, error) {
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
	return &httpContentRange{start, end, size}, nil
}