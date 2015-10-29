package mp4

import (
	"encoding/binary"
	"fmt"
	"io"
)

// UniBox - Universal not decoded Box
type UniBox struct {
	BoxHeader
	buff []byte
}

// DecodeUni - decodes uni
func DecodeUni(header BoxHeader) *UniBox {
	fmt.Printf("Universal not decoded box with header %#v\n", header)
	return &UniBox{
		BoxHeader: header,
	}
}

// Decode decodes a uni
func (u *UniBox) Decode(r io.Reader, size uint64) (Box, error) {
	u.buff = make([]byte, size)
	if err := binary.Read(r, binary.BigEndian, u.buff); err != nil {
		return nil, err
	}
	return u, nil
}

// Type returns teh real type of the box
func (u *UniBox) Type() string {
	return u.BoxHeader.Type
}

// Size returns the true size of the box
func (u *UniBox) Size() uint64 {
	return u.BoxHeader.Size
}

// Encode really encodes the box
func (u *UniBox) Encode(w io.Writer) error {
	if err := EncodeHeader(u, w); err != nil {
		return err
	}
	_, err := w.Write(u.buff)

	return err
}
