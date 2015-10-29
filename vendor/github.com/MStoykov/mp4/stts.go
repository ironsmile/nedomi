package mp4

import (
	"encoding/binary"
	"fmt"
	"io"
)

// Decoding Time to Sample Box (stts - mandatory)
//
// Contained in : Sample Table box (stbl)
//
// Status: decoded
//
// This table contains the duration in time units for each sample.
//
//   * sample count : the number of consecutive samples having the same duration
//   * time delta : duration in time units
type SttsBox struct {
	Version         byte
	Flags           [3]byte
	SampleCount     []uint32
	SampleTimeDelta []uint32
}

func DecodeStts(r io.Reader, size uint64) (Box, error) {
	data, err := read(r, size)

	if err != nil {
		return nil, err
	}

	c := binary.BigEndian.Uint32(data[4:8])
	b := &SttsBox{
		Flags:           [3]byte{data[1], data[2], data[3]},
		Version:         data[0],
		SampleCount:     make([]uint32, c),
		SampleTimeDelta: make([]uint32, c),
	}

	for i := 0; i < int(c); i++ {
		b.SampleCount[i] = binary.BigEndian.Uint32(data[(8 + 8*i):(12 + 8*i)])
		b.SampleTimeDelta[i] = binary.BigEndian.Uint32(data[(12 + 8*i):(16 + 8*i)])
	}

	return b, nil
}

func (b *SttsBox) Type() string {
	return "stts"
}

func (b *SttsBox) Size() uint64 {
	return uint64(8 + len(b.SampleCount)*8)
}

func (b *SttsBox) Dump() {
	fmt.Println("Time to sample:")
	for i := range b.SampleCount {
		fmt.Printf(" #%d : %d samples with duration %d units\n", i, b.SampleCount[i], b.SampleTimeDelta[i])
	}
}

func (b *SttsBox) Encode(w io.Writer) error {
	err := EncodeHeader(b, w)
	if err != nil {
		return err
	}
	buf := makebuf(b)
	buf[0] = b.Version
	buf[1], buf[2], buf[3] = b.Flags[0], b.Flags[1], b.Flags[2]
	binary.BigEndian.PutUint32(buf[4:], uint32(len(b.SampleCount)))
	for i := range b.SampleCount {
		binary.BigEndian.PutUint32(buf[8+8*i:], b.SampleCount[i])
		binary.BigEndian.PutUint32(buf[12+8*i:], b.SampleTimeDelta[i])
	}
	_, err = w.Write(buf)
	return err
}

// GetTimeCode returns the timecode (duration since the beginning of the media)
// of the beginning of a sample
func (b *SttsBox) GetTimeCode(sample uint32) (units uint32) {
	for i := 0; sample > 0 && i < len(b.SampleCount); i++ {
		if sample >= b.SampleCount[i] {
			units += b.SampleCount[i] * b.SampleTimeDelta[i]
			sample -= b.SampleCount[i]
		} else {
			units += sample * b.SampleTimeDelta[i]
			sample = 0
		}
	}

	return
}
