package storage

import (
	"bytes"
	"io"
	"io/ioutil"
	"testing"
)

func TestMultiReaderCloser(t *testing.T) {
	hello := bytes.NewBufferString("Hello")
	comma := bytes.NewBufferString(", ")
	world := bytes.NewBufferString("World!")
	mrc := newMultiReadCloser(ioutil.NopCloser(hello), ioutil.NopCloser(comma), ioutil.NopCloser(world))
	defer mrc.Close()
	var result bytes.Buffer
	result.ReadFrom(mrc)
	expected := "Hello, World!"
	if result.String() != expected {
		t.Fatalf("Expected to multiread [%s] read [%s]", expected, result.String())
	}
}

func TestLimitedReadCloser(t *testing.T) {
	hw := ioutil.NopCloser(bytes.NewBufferString("Hello, World!"))
	lrc := newLimitReadCloser(hw, 5)

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
	if err != io.EOF {
		t.Fatalf("expected EOF got %s", err)
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

	hw := ioutil.NopCloser(bytes.NewBufferString("Hello, World!"))
	src := newSkipReadCloser(hw, 5)
	defer src.Close()
	var result bytes.Buffer
	result.ReadFrom(src)
	expected := ", World!"
	if result.String() != expected {
		t.Fatalf("Expected to skipread [%s] read [%s]", expected, result.String())
	}
}
