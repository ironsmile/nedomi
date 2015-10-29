package mp4

import (
	"io"
)

// Tref Box (apple)
//
// Status: not decoded
type TrefBox struct {
	notDecoded []byte
}

func DecodeTref(r io.Reader, size uint64) (Box, error) {
	data, err := read(r, size)
	if err != nil {
		return nil, err
	}
	return &TrefBox{
		notDecoded: data[:],
	}, nil
}

func (b *TrefBox) Type() string {
	return "tref"
}

func (b *TrefBox) Size() uint64 {
	return uint64(len(b.notDecoded))
}

func (b *TrefBox) Encode(w io.Writer) error {
	err := EncodeHeader(b, w)
	if err != nil {
		return err
	}
	_, err = w.Write(b.notDecoded)
	return err
}
