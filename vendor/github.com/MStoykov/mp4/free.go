package mp4

import (
	"io"
)

// File Type Box (ftyp - mandatory)
//
// Status: decoded
type FreeBox struct {
	notDecoded []byte
}

func DecodeFree(r io.Reader, size uint64) (Box, error) {
	// !TODO check is seek is enough
	data, err := read(r, size)
	if err != nil {
		return nil, err
	}
	return &FreeBox{notDecoded: data}, nil
}

func (b *FreeBox) Type() string {
	return "free"
}

func (b *FreeBox) Size() uint64 {
	return uint64(len(b.notDecoded))
}

func (b *FreeBox) Encode(w io.Writer) error {
	err := EncodeHeader(b, w)
	if err != nil {
		return err
	}
	_, err = w.Write(b.notDecoded)
	return err
}
