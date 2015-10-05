package httputils

// This file has been based on http://golang.org/src/net/http/fs.go

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// Range specifies the byte range from a client request.
type Range struct {
	Start, Length uint64
}

// ContentRange returns Range as string appropriate for usage as value of Content-Range header.
func (r Range) ContentRange(size uint64) string {
	return fmt.Sprintf("bytes %d-%d/%d", r.Start, r.Start+r.Length-1, size)
}

// Range returns Range as string appropriate for usage as value of Range header.
func (r Range) Range() string {
	return fmt.Sprintf("bytes=%d-%d", r.Start, r.Start+r.Length-1)
}

// ParseRequestRange parses a client "Range" header string as per RFC 7233.
func ParseRequestRange(s string, size uint64) ([]Range, error) {
	if s == "" {
		return nil, nil // header not present
	}
	if size == 0 {
		return nil, errors.New("invalid size")
	}

	const b = "bytes="
	if !strings.HasPrefix(s, b) {
		return nil, errors.New("invalid range")
	}
	var ranges []Range
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
		r, err := parseReqByteRange(start, end, size)
		if err != nil {
			return nil, err
		}
		ranges = append(ranges, *r)
	}
	if len(ranges) < 1 {
		return nil, errors.New("invalid range")
	}
	return ranges, nil
}

// parseReqByteRange parses a byte range as per RFC 7233, section 2.1.
func parseReqByteRange(start, end string, size uint64) (*Range, error) {
	if start == "" {
		// If no start is specified, end specifies the
		// range starts relative to the end of the file.
		return parseEndRelativeRange(end, size)
	}

	si, err := strconv.ParseUint(start, 10, 64)
	if err != nil || si >= size {
		return nil, errors.New("invalid range")
	}
	if end == "" {
		// If no end is specified, range extends to end of the file.
		return &Range{Start: si, Length: size - si}, nil
	}

	ei, err := strconv.ParseUint(end, 10, 64)
	if err != nil || si > ei {
		return nil, errors.New("invalid range")
	}
	if ei >= size {
		ei = size - 1
	}
	return &Range{Start: si, Length: ei - si + 1}, nil
}

// parseEndRelativeRange handles ranges that use `suffix-byte-range-spec`
func parseEndRelativeRange(end string, size uint64) (*Range, error) {
	ei, err := strconv.ParseUint(end, 10, 64)
	if err != nil || ei == 0 {
		return nil, errors.New("invalid range")
	}
	if ei > size {
		ei = size
	}
	return &Range{Start: size - ei, Length: ei}, nil
}
