package utils

// This file has been based on http://golang.org/src/net/http/fs.go

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// HTTPRange specifies the byte range to be sent to the client.
type HTTPRange struct {
	Start, Length uint64
}

// ContentRange returns HTTPRange as string appropriate for usage as value of Content-Range  header
func (r HTTPRange) ContentRange(size uint64) string {
	return fmt.Sprintf("bytes %d-%d/%d", r.Start, r.Start+r.Length-1, size)
}

// parseReqByteRange parses a byte range as per RFC 7233, section 2.1.
func parseReqByteRange(start, end string, size uint64) (*HTTPRange, error) {
	if start == "" {
		// If no start is specified, end specifies the
		// range starts relative to the end of the file.
		ei, err := strconv.ParseUint(end, 10, 64)
		if err != nil || ei == 0 {
			return nil, errors.New("invalid range")
		}
		if ei > size {
			ei = size
		}
		return &HTTPRange{Start: size - ei, Length: ei}, nil
	}

	si, err := strconv.ParseUint(start, 10, 64)
	if err != nil || si >= size {
		return nil, errors.New("invalid range")
	}
	if end == "" {
		// If no end is specified, range extends to end of the file.
		return &HTTPRange{Start: si, Length: size - si}, nil
	}

	ei, err := strconv.ParseUint(end, 10, 64)
	if err != nil || si > ei {
		return nil, errors.New("invalid range")
	}
	if ei >= size {
		ei = size - 1
	}
	return &HTTPRange{Start: si, Length: ei - si + 1}, nil
}

// ParseReqRange parses a client "Range" header string as per RFC 7233.
func ParseReqRange(s string, size uint64) ([]HTTPRange, error) {
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
	var ranges []HTTPRange
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

// HTTPContentRange specifies the byte range to be sent to the client.
type HTTPContentRange struct {
	Start, Length, ObjSize uint64
}

// ParseRespContentRange parses a "Content-Range" header string. It only
// implements a subset of RFC 7233 - asterisks (unknown complete-length or
// unsatisfied-range) are treated as an error.
func ParseRespContentRange(cr string) (*HTTPContentRange, error) {
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
	return &HTTPContentRange{start, end - start + 1, size}, nil
}
