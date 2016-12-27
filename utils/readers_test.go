package utils

import (
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"testing"
	"testing/iotest"

	"github.com/ironsmile/nedomi/utils/testutils"
)

// errorReader returns the bytes provided but instead of io.EOF it returns the error provided
type errorReader struct {
	str []byte
	err error
}

func (e *errorReader) Read(p []byte) (n int, err error) {
	n = copy(p, e.str)
	e.str = e.str[n:]
	if len(e.str) == 0 {
		err = e.err
	}
	return
}

func newErrorReader(b []byte, e error) io.Reader {
	return &errorReader{
		str: b,
		err: e,
	}
}

func TestMultiReaderCloser(t *testing.T) {
	t.Parallel()
	hello := bytes.NewBufferString("Hello")
	comma := bytes.NewBufferString(", ")
	world := bytes.NewBufferString("World!")
	mrc := MultiReadCloser(ioutil.NopCloser(hello), ioutil.NopCloser(comma), ioutil.NopCloser(world))
	defer func() {
		if err := mrc.Close(); err != nil {
			t.Errorf(
				"MultiReadCloser.Close returned error - %s",
				err,
			)
		}
	}()
	var result bytes.Buffer
	if _, err := result.ReadFrom(mrc); err != nil {
		t.Fatalf("Unexpected ReadFrom error: %s", err)
	}
	expected := "Hello, World!"
	if result.String() != expected {
		t.Fatalf("Expected to multiread [%s] read [%s]", expected, result.String())
	}

}

func TestMultiReaderCloserReturnErrors(t *testing.T) {
	t.Parallel()
	hello := bytes.NewBufferString("Hello")
	comma := bytes.NewBufferString(", ")
	world := newErrorReader([]byte("Wor"), errors.New("testError"))
	mrc := MultiReadCloser(ioutil.NopCloser(hello), ioutil.NopCloser(comma), ioutil.NopCloser(world))
	defer func() {
		if err := mrc.Close(); err != nil {
			t.Errorf(
				"MultiReadCloser.Close returned error - %s",
				err,
			)
		}
	}()
	var result bytes.Buffer
	if _, err := result.ReadFrom(mrc); err == nil || err.Error() != "testError" {
		t.Fatalf("Unexpected ReadFrom error: '%s' expected to get error 'testError'", err)
	}
	expected := "Hello, Wor"
	if result.String() != expected {
		t.Fatalf("Expected to multiread [%s] read [%s]", expected, result.String())
	}
}

func TestMultiReaderCloserWriteTo(t *testing.T) {
	t.Parallel()
	hello := bytes.NewBufferString("Hello")
	comma := bytes.NewBufferString(", ")
	world := bytes.NewBufferString("World!")
	mrc := MultiReadCloser(ioutil.NopCloser(hello), ioutil.NopCloser(comma), ioutil.NopCloser(world))
	defer func() {
		if err := mrc.Close(); err != nil {
			t.Errorf(
				"MultiReadCloser.Close returned error - %s",
				err,
			)
		}
	}()
	var result bytes.Buffer
	if _, err := mrc.(io.WriterTo).WriteTo(&result); err != nil {
		t.Fatalf("Unexpected ReadFrom error: %s", err)
	}
	expected := "Hello, World!"
	if result.String() != expected {
		t.Fatalf("Expected to multiread [%s] read [%s]", expected, result.String())
	}
}

func TestMultiReaderCloserWriteToReturnErrors(t *testing.T) {
	t.Parallel()
	hello := bytes.NewBufferString("Hello")
	comma := bytes.NewBufferString(", ")
	world := newErrorReader([]byte("Wor"), errors.New("testError"))
	mrc := MultiReadCloser(ioutil.NopCloser(hello), ioutil.NopCloser(comma), ioutil.NopCloser(world))
	defer func() {
		if err := mrc.Close(); err != nil {
			t.Errorf(
				"MultiReadCloser.Close returned error - %s",
				err,
			)
		}
	}()
	var result bytes.Buffer
	if _, err := mrc.(io.WriterTo).WriteTo(&result); err == nil || err.Error() != "testError" {
		t.Fatalf("Unexpected ReadFrom error: '%s' expected to get error 'testError'", err)
	}
	expected := "Hello, Wor"
	if result.String() != expected {
		t.Fatalf("Expected to multiread [%s] read [%s]", expected, result.String())
	}
}

func TestMultiReaderCloserNoReaders(t *testing.T) {
	t.Parallel()
	mrc := MultiReadCloser()
	defer func() {
		if err := mrc.Close(); err != nil {
			t.Errorf(
				"MultiReadCloser.Close returned error - %s",
				err,
			)
		}
	}()
	var result bytes.Buffer
	if n, err := result.ReadFrom(mrc); err != nil {
		t.Fatalf("Unexpected ReadFrom error: %s, expected nil", err)
	} else if n != 0 {
		t.Fatalf("Expected to read 0 bytes from empty MultiReader, but read %d", n)

	}
	expected := ""
	if result.String() != expected {
		t.Fatalf("Expected to multiread [%s] read [%s]", expected, result.String())
	}
}

func TestMultiReaderCloserNoReadersWriteTo(t *testing.T) {
	t.Parallel()
	mrc := MultiReadCloser()
	defer func() {
		if err := mrc.Close(); err != nil {
			t.Errorf(
				"MultiReadCloser.Close returned error - %s",
				err,
			)
		}
	}()
	var result bytes.Buffer
	if n, err := mrc.(io.WriterTo).WriteTo(&result); err != nil {
		t.Fatalf("Unexpected ReadFrom error: %s, expected nil", err)
	} else if n != 0 {
		t.Fatalf("Expected to read 0 bytes from empty MultiReader, but read %d", n)

	}
	expected := ""
	if result.String() != expected {
		t.Fatalf("Expected to multiread [%s] read [%s]", expected, result.String())
	}
}
func TestLimitedReadCloser(t *testing.T) {
	t.Parallel()
	hw := ioutil.NopCloser(bytes.NewBufferString("Hello, World!"))
	lrc := LimitReadCloser(hw, 5)

	var p [10]byte
	size, err := lrc.Read(p[:2])
	if err != nil {
		t.Fatal(err)
	} else if size != 2 {
		t.Fatalf("expected to read 2 from limitReader but read %d", size)
	}
	size, err = lrc.Read(p[2:4])
	if err != nil {
		t.Fatal(err)
	} else if size != 2 {
		t.Fatalf("expected to read 2 from limitReader but read %d", size)
	}
	size, err = lrc.Read(p[4:6])
	if err != io.EOF && err != nil {
		t.Fatalf("expected EOF or nil got %s", err)
	} else if size != 1 {
		t.Fatalf("expected to read 1 from limitReader but read %d", size)
	}
	size, err = lrc.Read(p[6:8])
	if err != io.EOF {
		t.Fatalf("expected EOF got %s", err)
	} else if size != 0 {
		t.Fatalf("expected to read 0 from limitReader but read %d", size)
	}

	var expected = [10]byte{}
	copy(expected[:], []byte("Hello"))
	if !bytes.Equal(p[:], expected[:]) {
		t.Fatalf("Expected to have read [%s] but read [%s]", expected, p)

	}
}

func TestSkipReaderClose(t *testing.T) {
	t.Parallel()
	hw := ioutil.NopCloser(bytes.NewBufferString("Hello, World!"))
	src, err := SkipReadCloser(hw, 5)
	if err != nil {
		t.Fatal("unexpected error", err)
	}
	defer func() {
		testutils.ShouldntFail(t, src.Close())
	}()
	var result bytes.Buffer
	if _, err := result.ReadFrom(src); err != nil {
		t.Fatalf("Unexpected ReadFrom error: %s", err)
	}
	expected := ", World!"
	if result.String() != expected {
		t.Fatalf("Expected to skipread [%s] read [%s]", expected, result.String())
	}
}

func TestSkipReaderCloseWithPipe(t *testing.T) {
	t.Parallel()
	var input = []byte{'a', 'b', 'c', 'd'}
	var output = []byte{'b', 'c', 'd'}
	r, w := io.Pipe()
	src, err := SkipReadCloser(r, 1)
	if err != nil {
		t.Fatal("unexpected error", err)
	}
	go func() {
		if _, err := w.Write(input); err != nil {
			t.Fatalf("Unexpected Write error: %s", err)
		}
		testutils.ShouldntFail(t, w.Close())
	}()
	defer func() {
		testutils.ShouldntFail(t, src.Close())
	}()

	var result bytes.Buffer
	if _, err := result.ReadFrom(iotest.OneByteReader(src)); err != nil {
		t.Fatalf("Unexpected ReadFrom error: %s", err)
	}
	if result.String() != string(output) {
		t.Fatalf("Expected to skipread [%s] read [%s]", output, result.String())
	}
}
