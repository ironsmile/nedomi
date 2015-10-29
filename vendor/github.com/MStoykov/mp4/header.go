package mp4

import (
	"encoding/binary"
	"io"
	"math"
)

const (
	// BoxHeaderSize the size of a box
	boxHeaderSize32bit = 8
	boxHeaderSize64bit = 16
)

// BoxHeader The header of a box
type BoxHeader struct {
	Type string
	Size uint64
}

// DecodeHeader decodes a box header (size + box type)
func DecodeHeader(r io.Reader) (BoxHeader, error) {
	buf := make([]byte, boxHeaderSize32bit)
	n, err := r.Read(buf)
	if err != nil {
		return BoxHeader{}, err
	}
	if n != boxHeaderSize32bit {
		return BoxHeader{}, ErrTruncatedHeader
	}
	typeName := string(buf[4:8])
	size := uint64(binary.BigEndian.Uint32(buf[0:4]))
	if size == 1 { // 64 bit size
		buf = make([]byte, boxHeaderSize64bit-boxHeaderSize32bit)
		n, err := r.Read(buf)
		if err != nil {
			return BoxHeader{}, err
		}
		if n != 8 {
			return BoxHeader{}, ErrTruncatedHeader
		}
		size = binary.BigEndian.Uint64(buf[0:8])
	}
	return BoxHeader{typeName, size}, nil
}

// EncodeHeader encodes a box header to a writer
func EncodeHeader(b Box, w io.Writer) error {
	var size = b.Size()
	size = AddHeaderSize(size)
	if math.MaxUint32 > size {
		buf := make([]byte, boxHeaderSize32bit)
		binary.BigEndian.PutUint32(buf, uint32(size))
		strtobuf(buf[4:], b.Type(), 4)
		_, err := w.Write(buf)
		return err
	}
	buf := make([]byte, boxHeaderSize64bit)
	binary.BigEndian.PutUint32(buf, uint32(1))
	strtobuf(buf[4:], b.Type(), 4)
	binary.BigEndian.PutUint64(buf[8:], size)
	_, err := w.Write(buf)
	return err
}

func AddHeaderSize(size uint64) uint64 {
	return size + HeaderSizeFor(size)
}

func RemoveHeaderSize(size uint64) uint64 {
	if size > math.MaxUint32 {
		return size - boxHeaderSize64bit
	}
	return size - boxHeaderSize32bit
}

func HeaderSizeFor(size uint64) uint64 {
	if math.MaxUint32 > size+8 {
		return boxHeaderSize32bit
	}
	return boxHeaderSize64bit
}
