package mp4

import (
	"encoding/binary"
	"io"
)

// CttsBox - Composition Time to Sample Box (ctts - optional)
//
// Contained in: Sample Table Box (stbl)
//
// Status: version 0 decoded. version 1 uses int32 for offsets
type CttsBox struct {
	Version      byte
	Flags        [3]byte
	SampleCount  []uint32
	SampleOffset []uint32 // int32 for version 1
}

func DecodeCtts(r io.Reader, size uint64) (Box, error) {
	data, err := read(r, size)

	if err != nil {
		return nil, err
	}

	c := binary.BigEndian.Uint32(data[4:8])
	b := &CttsBox{
		Flags:        [3]byte{data[1], data[2], data[3]},
		Version:      data[0],
		SampleCount:  make([]uint32, c),
		SampleOffset: make([]uint32, c),
	}

	for i := 0; i < int(c); i++ {
		b.SampleCount[i] = binary.BigEndian.Uint32(data[(8 + 8*i):(12 + 8*i)])
		b.SampleOffset[i] = binary.BigEndian.Uint32(data[(12 + 8*i):(16 + 8*i)])
	}

	return b, nil
}

// Type returns ctts
func (b *CttsBox) Type() string {
	return "ctts"
}

// Size returns the size of ctts
func (b *CttsBox) Size() uint64 {
	return uint64(8 + len(b.SampleCount)*8)
}

// Encode encodes
func (b *CttsBox) Encode(w io.Writer) error {
	if err := EncodeHeader(b, w); err != nil {
		return err
	}
	buf := makebuf(b)
	buf[0] = b.Version
	buf[1], buf[2], buf[3] = b.Flags[0], b.Flags[1], b.Flags[2]
	binary.BigEndian.PutUint32(buf[4:], uint32(len(b.SampleCount)))
	for i := range b.SampleCount {
		binary.BigEndian.PutUint32(buf[8+8*i:], b.SampleCount[i])
		binary.BigEndian.PutUint32(buf[12+8*i:], b.SampleOffset[i])
	}
	_, err := w.Write(buf)
	return err
}
