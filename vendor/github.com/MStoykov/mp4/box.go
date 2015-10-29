package mp4

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

var (
	// ErrUnknownBoxType is for unknown box types
	ErrUnknownBoxType = errors.New("unknown box type")
	// ErrTruncatedHeader is when a head gets truncated
	ErrTruncatedHeader = errors.New("truncated header")
	errSmallRead       = errors.New("read less than expected")
)

// BadFormatErr is type of error when an unexpected box appears in unexpected places
type BadFormatErr struct {
	enclosingBox, unexpectedBox string
}

func (b *BadFormatErr) Error() string {
	return fmt.Sprintf("Bad format: unexpected %s box inside box %s",
		b.unexpectedBox, b.enclosingBox)

}

var decoders map[string]boxDecoder

func init() {
	decoders = map[string]boxDecoder{
		"ftyp": DecodeFtyp,
		"moov": DecodeMoov,
		"mvhd": DecodeMvhd,
		"iods": DecodeIods,
		"trak": DecodeTrak,
		"udta": DecodeUdta,
		"tkhd": DecodeTkhd,
		"edts": DecodeEdts,
		"elst": DecodeElst,
		"mdia": DecodeMdia,
		"minf": DecodeMinf,
		"mdhd": DecodeMdhd,
		"hdlr": DecodeHdlr,
		"vmhd": DecodeVmhd,
		"smhd": DecodeSmhd,
		"dinf": DecodeDinf,
		"dref": DecodeDref,
		"sbgp": DecodeSbgp,
		"sdtp": DecodeSdtp,
		"sgpd": DecodeSgpd,
		"stbl": DecodeStbl,
		"stco": DecodeStco,
		"stsc": DecodeStsc,
		"stsz": DecodeStsz,
		"ctts": DecodeCtts,
		"stsd": DecodeStsd,
		"stts": DecodeStts,
		"stss": DecodeStss,
		"meta": DecodeMeta,
		"mdat": DecodeMdat,
		"free": DecodeFree,
		"name": DecodeName,
		"tref": DecodeTref,
		"gmhd": DecodeGmhd,
		"chpl": DecodeChpl,
		"co64": DecodeCo64,
	}
}

// Box an atom
type Box interface {
	Type() string
	Size() uint64
	Encode(w io.Writer) error
}

type boxDecoder func(r io.Reader, size uint64) (Box, error)

// DecodeBox decodes a box
func DecodeBox(h BoxHeader, r io.Reader) (Box, error) {
	d := decoders[h.Type]
	if d == nil {
		d = DecodeUni(h).Decode
	}
	b, err := d(r, RemoveHeaderSize(h.Size))
	if err != nil {
		return nil, err
	}
	return b, nil
}

// DecodeContainer decodes a container box
func DecodeContainer(r io.Reader, size uint64) ([]Box, error) {
	l := []Box{}
	for {
		h, err := DecodeHeader(r)
		if err == io.EOF {
			return l, nil
		}
		if err != nil {
			return l, err
		}
		b, err := DecodeBox(h, r)
		if err != nil {
			return l, err
		}
		l = append(l, b)
		size -= h.Size
		if size == 0 {
			return l, nil
		}
	}
}

// Fixed16 is 8.8 fixed point number
type Fixed16 uint16

func (f Fixed16) String() string {
	return fmt.Sprintf("%d.%d", uint16(f)>>8, uint16(f)&7)
}

func fixed16(bytes []byte) Fixed16 {
	return Fixed16(binary.BigEndian.Uint16(bytes))
}

func putFixed16(bytes []byte, i Fixed16) {
	binary.BigEndian.PutUint16(bytes, uint16(i))
}

// Fixed32 is 16.16 fixed point number
type Fixed32 uint32

func (f Fixed32) String() string {
	return fmt.Sprintf("%d.%d", uint32(f)>>16, uint32(f)&15)
}

func fixed32(bytes []byte) Fixed32 {
	return Fixed32(binary.BigEndian.Uint32(bytes))
}

func putFixed32(bytes []byte, i Fixed32) {
	binary.BigEndian.PutUint32(bytes, uint32(i))
}

func strtobuf(out []byte, str string, l int) {
	in := []byte(str)
	if l < len(in) {
		copy(out, in)
	} else {
		copy(out, in[0:l])
	}
}
func makebuf(b Box) []byte {
	return make([]byte, b.Size())
}

func read(r io.Reader, size uint64) ([]byte, error) {
	var buf = make([]byte, size)
	if readSize, err := io.ReadFull(r, buf); err != nil && err != io.EOF {
		return nil, err
	} else if readSize != int(size) {
		return nil, fmt.Errorf("read %d which is different from the expected %d", readSize, size)
	}
	return buf, nil
}
