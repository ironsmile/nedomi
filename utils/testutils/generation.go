package testutils

import (
	"encoding/binary"
	"encoding/hex"
	"io"
	"math/rand"
)

// GenerateMeAString generates a pseudo-randomized string of the provided size, with the
func GenerateMeAString(seed, size int64) string {
	var b = make([]byte, size/2+size%2) // hex encoding produce twice the amount of bytes it consumes
	r := readerFromSource(rand.NewSource(seed))
	if _, err := r.Read(b); err != nil {
		panic(err)
	}

	return hex.EncodeToString(b)[:size] // odd cases
}

type sourceReader struct {
	rand.Source
}

func (s *sourceReader) Read(b []byte) (int, error) {
	l := len(b)
	var buf [8]byte
	for a := 0; l-1 > a; a += 8 {
		binary.LittleEndian.PutUint64(buf[:], uint64(s.Int63()))
		copy(b[a:], buf[:])
	}
	return l, nil
}

func readerFromSource(s rand.Source) io.Reader {
	return &sourceReader{s}
}
