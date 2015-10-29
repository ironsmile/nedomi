package mp4

import (
	"io"
)

// Sbgp Box
//
// Status: not decoded
type SbgpBox struct {
	notDecoded []byte
}

func DecodeSbgp(r io.Reader, size uint64) (Box, error) {
	data, err := read(r, size)
	if err != nil {
		return nil, err
	}
	return &SbgpBox{
		notDecoded: data[:],
	}, nil
}

func (b *SbgpBox) Type() string {
	return "sbgp"
}

func (b *SbgpBox) Size() uint64 {
	return uint64(len(b.notDecoded))
}

func (b *SbgpBox) Encode(w io.Writer) error {
	err := EncodeHeader(b, w)
	if err != nil {
		return err
	}
	_, err = w.Write(b.notDecoded)
	return err
}
