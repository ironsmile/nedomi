package utils

import (
	"bytes"
	"io"
	"reflect"
	"testing"
)

var data = []byte("some bytes for testing the awesomest test")

func testReader() io.Reader {
	return bytes.NewReader(data)
}

func TestCopyN(t *testing.T) {
	var tests = []struct {
		reader   io.Reader
		limit    int64
		expected int64
	}{
		{io.LimitReader(testReader(), 2), 1, 1},
		{io.LimitReader(testReader(), 1), 2, 1},
		{io.LimitReader(testReader(), 2), 2, 2},
		{io.LimitReader(io.LimitReader(testReader(), 1), 2), 3, 1},
		{io.LimitReader(io.LimitReader(testReader(), 2), 1), 3, 1},
		{io.LimitReader(io.LimitReader(testReader(), 3), 2), 1, 1},
		{io.LimitReader(io.LimitReader(testReader(), 3), 2), 2, 2},
		{io.LimitReader(io.LimitReader(io.LimitReader(testReader(), 1), 3), 2), 2, 1},
		{io.LimitReader(io.LimitReader(io.LimitReader(testReader(), 2), 3), 2), 2, 2},
		{io.LimitReader(io.LimitReader(io.LimitReader(testReader(), 3), 3), 5), 5, 3},
		{io.LimitReader(io.LimitReader(io.LimitReader(testReader(), 4), 3), 2), 2, 2},
		{io.LimitReader(io.LimitReader(io.LimitReader(io.LimitReader(testReader(), 300), 400), 300), 100), 200, int64(len(data))},
	}

	for index, test := range tests {
		var buf = new(bytes.Buffer)
		n, err := CopyN(buf, test.reader, test.limit)
		if err != nil {
			t.Fatalf("test %d returned %d and %s `%s`", index, n, err, buf)
		}
		if n != test.expected {
			t.Errorf("it was expected test %d to copy %d but it copied %d",
				index, test.expected, n)
		}
		if test.expected != int64(buf.Len()) {
			t.Errorf("it was expected test %d to write %d to buffer not %d",
				index, test.expected, buf.Len())
		}
		if !reflect.DeepEqual(data[:test.expected], buf.Bytes()) {
			t.Errorf("it was expected test %d to have in buffers \n%s\nnot\n%s",
				index, data[:test.expected], buf.Bytes())
		}
	}
}
