package throttle

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"math/rand"
	"testing"
	"time"
)

const (
	testMinWrite = 8 * 1024
)

type testFunc func(testing.TB, throttleTest)

var content = map[string][]byte{
	"20M": generate(1024 * 1024 * 20),
	"2M":  generate(1024 * 1024 * 2),
	"10K": generate(1024 * 10),
	"1K":  generate(1024),
}

func generate(n int64) []byte {
	var result = make([]byte, n)
	var r = rand.New(rand.NewSource(n))
	for i := int64(0); n > i+8; i++ {
		binary.BigEndian.PutUint64(result[i:], uint64(r.Int63()))
	}
	return result
}

type throttleTest struct {
	info  string
	data  []byte
	speed int64
}

var tests = [...]throttleTest{
	{"20M with 1MB/s", content["20M"], 1024 * 1024},
	{"2M with 200KB/s", content["2M"], 200 * 1024},
	{"2M with 1MB/s", content["2M"], 1024 * 1024},
	{"10K with 1MB/s", content["10K"], 1024 * 1024},
	{"1K with 100B/s", content["1K"], 100},
}

func TestResponseWriter(t *testing.T) {
	t.Parallel()
	runInParallel(t, tests[:], testResponseWriter)
}

func testResponseWriter(t testing.TB, test throttleTest) {
	var ew = &expectantWriter{expects: test.data}
	var tw = NewThrottleWriter(ew, test.speed, testMinWrite)
	testCopy(t, tw, bytes.NewReader(test.data), int64(len(test.data)), test.speed, test.info)
}

func TestResponseWriterWithReadFrom(t *testing.T) {
	t.Parallel()
	runInParallel(t, tests[:], testResponseWriterWithReadFrom)
}

func testResponseWriterWithReadFrom(t testing.TB, test throttleTest) {
	var ew = &expectantWriter{expects: test.data}
	var tw = NewThrottleWriter(ew, test.speed, testMinWrite)
	testCopy(t, tw,
		reader{bytes.NewReader(test.data)},
		int64(len(test.data)),
		test.speed,
		test.info,
	)
}

func runInParallel(t testing.TB, tests []throttleTest, testIt testFunc) {
	var chs []chan struct{}
	for _, test := range tests {
		var ch = make(chan struct{})
		chs = append(chs, ch)
		go func(test throttleTest) {
			defer close(ch)
			testIt(t, test)
		}(test)
	}
	for _, ch := range chs {
		<-ch
	}
}

func around(l, r time.Duration) bool {
	var ln, rn = l.Nanoseconds(), r.Nanoseconds()
	var div = abs(ln - rn)
	return ln/10 > div || rn/10 > div || int64(time.Second/2) > div
}

func abs(a int64) int64 {
	if a > 0 {
		return a
	}
	return -a
}

type expectantWriter struct {
	expects []byte
	n       int
}

func (e *expectantWriter) Write(b []byte) (n int, err error) {
	for _, aByte := range b {
		if e.expects[e.n] != aByte {
			return n, fmt.Errorf(
				"the %dth symbol was supposed to be %c but was %c",
				e.n, e.expects[e.n], aByte)
		}
		e.n++
		n++
	}
	return n, nil
}

type reader struct {
	io.Reader
}

type expectantReaderFrom struct {
	expects []byte
	n       int64
}

func (e *expectantReaderFrom) ReadFrom(r io.Reader) (n int64, err error) {
	var b = make([]byte, 1024)
	var nn = 0
	for ; err == nil; n += int64(nn) {
		nn, err = r.Read(b[:])
		for _, aByte := range b[:nn] {
			if e.expects[e.n] != aByte {
				return n, fmt.Errorf(
					"the %dth symbol was supposed to be %c but was %c",
					e.n, e.expects[e.n], aByte)
			}
			e.n++
		}
	}
	if err == io.EOF {
		err = nil
	}
	return
}

func testCopy(
	t testing.TB,
	w io.Writer,
	r io.Reader,
	expectedSize, expectedSpeed int64,
	info string,
) {
	now := time.Now()
	n, err := io.Copy(w, r)
	end := time.Now()
	if err != nil {
		t.Fatalf("[%s]unexpected error from io.Copy :%s", info, err)
	}
	if n != expectedSize {
		t.Errorf("[%s]the expected size of copy was %d but it was %d",
			info, expectedSize, n)
	}
	got := end.Sub(now)
	expected := time.Duration((n / expectedSpeed)) * time.Second
	if !around(got, expected) {
		t.Errorf("[%s]the throttledWriting took %s which isn't around %s",
			info, got, expected)
	}
}
