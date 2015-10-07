package cache

import (
	"io"
	"io/ioutil"
	"testing"
)

type countingReader int

func (c *countingReader) Read(b []byte) (int, error) {
	*c += countingReader(len(b))
	return len(b), nil
}

func read(t *testing.T, r io.Reader, b []byte) {
	n, err := r.Read(b)
	if err != nil {
		t.Errorf("error reading %d bytes from %+v - %s",
			len(b), r, err)
	}
	if n != len(b) {
		t.Errorf("read %d bytes from %+v  instead of %d",
			n, r, len(b))
	}
}

func TestPartReaderCloser(t *testing.T) {
	var cr = countingReader(0)
	var r = ioutil.NopCloser(&cr)
	var chr = newWholeChunkReadCloser(r, 4)
	var b [20]byte
	read(t, chr, b[:3])
	if cr != 3 {
		t.Errorf("expected to have read %d not %d", 3, cr)
	}
	read(t, chr, b[:2])
	if cr != 5 {
		t.Errorf("expected to have read %d not %d", 5, cr)
	}
	err := chr.Close()
	if err != nil {
		t.Errorf("unexpected error %s", err)
	}
	if cr != 8 {
		t.Errorf("expected to have read %d not %d", 8, cr)
	}
}
func TestPartReaderCloserExact(t *testing.T) {
	var cr = countingReader(0)
	var r = ioutil.NopCloser(&cr)
	var chr = newWholeChunkReadCloser(r, 4)
	var b [20]byte
	read(t, chr, b[:3])
	if cr != 3 {
		t.Errorf("expected to have read %d not %d", 3, cr)
	}
	read(t, chr, b[:2])
	if cr != 5 {
		t.Errorf("expected to have read %d not %d", 5, cr)
	}
	read(t, chr, b[:3])
	if cr != 8 {
		t.Errorf("expected to have read %d not %d", 8, cr)
	}
	err := chr.Close()
	if err != nil {
		t.Errorf("unexpected error %s", err)
	}
	if cr != 8 {
		t.Errorf("expected to have read %d not %d", 8, cr)
	}
}

func TestPartReaderCloserNoReads(t *testing.T) {
	var cr = countingReader(0)
	var r = ioutil.NopCloser(&cr)
	var chr = newWholeChunkReadCloser(r, 4)
	err := chr.Close()
	if err != nil {
		t.Errorf("unexpected error %s", err)
	}
	if cr != 0 {
		t.Errorf("expected to have read %d not %d", 0, cr)
	}
}
