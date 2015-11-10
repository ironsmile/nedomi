package mp4

import (
	"fmt"
	"io"
)

// FtypBox - File Type Box (ftyp - mandatory)
//
// Status: decoded
type FtypBox struct {
	MajorBrand       string
	MinorVersion     []byte
	CompatibleBrands []string
}

// Decode decodes the ftyp box
func DecodeFtyp(r io.Reader, size uint64) (Box, error) {
	data, err := read(r, size)
	if err != nil {
		return nil, err
	}
	b := &FtypBox{
		MajorBrand:   string(data[0:4]),
		MinorVersion: data[4:8],
	}
	if len(data) > 8 {
		b.CompatibleBrands = make([]string, (len(data)-8)/4)
		for i := 8; i < len(data); i += 4 {
			b.CompatibleBrands[(i-8)/4] = string(data[i : i+4])
		}
	}
	return b, nil
}

func (b *FtypBox) Type() string {
	return "ftyp"
}

func (b *FtypBox) Size() uint64 {
	return uint64(8 + 4*len(b.CompatibleBrands))
}

func (b *FtypBox) Dump() {
	fmt.Printf("File Type: %s\n", b.MajorBrand)
}

func (b *FtypBox) Encode(w io.Writer) error {
	err := EncodeHeader(b, w)
	if err != nil {
		return err
	}
	buf := makebuf(b)
	strtobuf(buf, b.MajorBrand, 4)
	copy(buf[4:], b.MinorVersion)
	for i, c := range b.CompatibleBrands {
		strtobuf(buf[8+i*4:], c, 4)
	}
	_, err = w.Write(buf)
	return err
}
