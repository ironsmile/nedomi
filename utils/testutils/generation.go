package testutils

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"math/rand"
	"net/url"

	"github.com/ironsmile/nedomi/types"
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

// GetUpstream returns a fully configured upstream address
func GetUpstream(i int) *types.UpstreamAddress {
	return &types.UpstreamAddress{
		URL:         url.URL{Host: fmt.Sprintf("127.0.%d.%d", (i/256)%256, i%256), Scheme: "http"},
		Hostname:    fmt.Sprintf("www.upstream%d.com", i),
		Port:        "80",
		OriginalURL: &url.URL{Host: fmt.Sprintf("www.upstream%d.com", i), Scheme: "http"},
		Weight:      100 + uint32(rand.Intn(500)),
	}
}

// GetUpstreams returns a fully configured slice of sequential upstream address
func GetUpstreams(from, to int) []*types.UpstreamAddress {
	result := make([]*types.UpstreamAddress, to-from+1)
	for i := from; i <= to; i++ {
		result[i-from] = GetUpstream(i)
	}
	return result
}

// GetRandomUpstreams returns a slice with a random number of upstreams
func GetRandomUpstreams(minCount, maxCount int) []*types.UpstreamAddress {
	count := minCount + rand.Intn(maxCount-minCount)
	return GetUpstreams(1, count)
}
