package utils

import (
	"bytes"
	"io"
	"math/rand"
	"testing"
)

func TestMultiWriterWithNWriters(t *testing.T) {
	t.Parallel()
	var writers = make([]io.WriteCloser, rand.Intn(20))
	for index := range writers {
		writers[index] = NopCloser(new(bytes.Buffer))
	}
	var multi = MultiWriteCloser(writers...)
	var expected = []byte(`Hello, World!`)
	multi.Write(expected[0:5])
	multi.Write(expected[5:8])
	multi.Write(expected[8:])
	for index, writer := range writers {
		got := (unwrapNopCloser(writer)).(interface {
			Bytes() []byte
		}).Bytes()
		if string(got) != string(expected) {
			t.Errorf("writer %d got `%+v` not `%+v`", index, got, expected)
		}
	}
}

func unwrapNopCloser(input io.Writer) io.Writer {
	return input.(nopCloser).Writer
}
