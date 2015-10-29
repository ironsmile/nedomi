package mp4

import (
	"io"
)

// Name Box
//
// Status: not decoded
type NameBox struct {
	notDecoded []byte
}

func DecodeName(r io.Reader, size uint64) (Box, error) {
	data, err := read(r, size)
	if err != nil {
		return nil, err
	}
	return &NameBox{
		notDecoded: data[:],
	}, nil
}

func (b *NameBox) Type() string {
	return "name"
}

func (b *NameBox) Size() uint64 {
	return uint64(len(b.notDecoded))
}

func (b *NameBox) Encode(w io.Writer) error {
	err := EncodeHeader(b, w)
	if err != nil {
		return err
	}
	_, err = w.Write(b.notDecoded)
	return err
}
