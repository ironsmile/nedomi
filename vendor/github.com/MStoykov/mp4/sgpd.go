package mp4

import (
	"io"
)

// Sgpd Box
//
// Status: not decoded
type SgpdBox struct {
	notDecoded []byte
}

func DecodeSgpd(r io.Reader, size uint64) (Box, error) {
	data, err := read(r, size)
	if err != nil {
		return nil, err
	}
	return &SgpdBox{
		notDecoded: data[:],
	}, nil
}

func (b *SgpdBox) Type() string {
	return "sgpd"
}

func (b *SgpdBox) Size() uint64 {
	return uint64(len(b.notDecoded))
}

func (b *SgpdBox) Encode(w io.Writer) error {
	err := EncodeHeader(b, w)
	if err != nil {
		return err
	}
	_, err = w.Write(b.notDecoded)
	return err
}
