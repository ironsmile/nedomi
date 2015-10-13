package cache

import (
	"io/ioutil"
	"testing"
)

type countingReader int

func (c *countingReader) Read(b []byte) (int, error) {
	*c += countingReader(len(b))
	return len(b), nil
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
