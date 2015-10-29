package mp4

import (
	"encoding/binary"
	"fmt"
	"io"
)

// ElstBox - Edit List Box (elst - optional)
//
// Contained in : Edit Box (edts)
//
// Status: version 0 decoded. version 1 not supported
type ElstBox struct {
	Version                             byte
	Flags                               [3]byte
	SegmentDuration, MediaTime          []uint32 // should be uint32/int32 for version 0 and uint64/int32 for version 1
	MediaRateInteger, MediaRateFraction []uint16 // should be int16
}

func DecodeElst(r io.Reader, size uint64) (Box, error) {
	data, err := read(r, size)
	if err != nil {
		return nil, err
	}

	c := binary.BigEndian.Uint32(data[4:8])
	b := &ElstBox{
		Flags:             [3]byte{data[1], data[2], data[3]},
		Version:           data[0],
		MediaTime:         make([]uint32, c),
		SegmentDuration:   make([]uint32, c),
		MediaRateInteger:  make([]uint16, c),
		MediaRateFraction: make([]uint16, c),
	}

	for i := 0; i < int(c); i++ {
		b.MediaTime[i] = binary.BigEndian.Uint32(data[(12 + 12*i):(16 + 12*i)])
		b.SegmentDuration[i] = binary.BigEndian.Uint32(data[(8 + 12*i):(12 + 12*i)])
		b.MediaRateInteger[i] = binary.BigEndian.Uint16(data[(16 + 12*i):(18 + 12*i)])
		b.MediaRateFraction[i] = binary.BigEndian.Uint16(data[(18 + 12*i):(20 + 12*i)])
	}

	return b, nil
}

// Type returns elst
func (b *ElstBox) Type() string {
	return "elst"
}

// Size retruns the size of ElstBox
func (b *ElstBox) Size() uint64 {
	return uint64(8 + len(b.SegmentDuration)*12)
}

// Dump dumps the elst box
func (b *ElstBox) Dump() {
	fmt.Println("Segment Duration:")
	for i, d := range b.SegmentDuration {
		fmt.Printf(" #%d: %d units\n", i, d)
	}
}

// Encode encodes the elst box
func (b *ElstBox) Encode(w io.Writer) error {
	err := EncodeHeader(b, w)
	if err != nil {
		return err
	}
	buf := makebuf(b)
	buf[0] = b.Version
	buf[1], buf[2], buf[3] = b.Flags[0], b.Flags[1], b.Flags[2]
	binary.BigEndian.PutUint32(buf[4:], uint32(len(b.SegmentDuration)))
	for i := range b.SegmentDuration {
		binary.BigEndian.PutUint32(buf[8+12*i:], b.SegmentDuration[i])
		binary.BigEndian.PutUint32(buf[12+12*i:], b.MediaTime[i])
		binary.BigEndian.PutUint16(buf[16+12*i:], b.MediaRateInteger[i])
		binary.BigEndian.PutUint16(buf[18+12*i:], b.MediaRateFraction[i])
	}
	_, err = w.Write(buf)
	return err
}
