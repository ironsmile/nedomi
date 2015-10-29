package mp4

import (
	"encoding/binary"
	"fmt"
	"io"
)

// StcoBox is Chunk Offset Box (stco - mandatory)
//
// Contained in : Sample Table box (stbl)
//
// Status: decoded
//
// This is the 32bits version of the box, the 64bits version (co64) is not decoded.
//
// The table contains the offsets (starting at the beginning of the file) for each chunk of data for the current track.
// A chunk contains samples, the table defining the allocation of samples to each chunk is stsc.
type StcoBox struct {
	Version     byte
	Flags       [3]byte
	ChunkOffset []uint32
}

// DecodeStco does what it says on the tin
func DecodeStco(r io.Reader, size uint64) (Box, error) {
	data, err := read(r, size)

	if err != nil {
		return nil, err
	}

	c := binary.BigEndian.Uint32(data[4:8])
	b := &StcoBox{
		Flags:       [3]byte{data[1], data[2], data[3]},
		Version:     data[0],
		ChunkOffset: make([]uint32, c),
	}

	for i := 0; i < int(c); i++ {
		b.ChunkOffset[i] = binary.BigEndian.Uint32(data[(8 + 4*i):(12 + 4*i)])
	}

	return b, nil
}

// Type returns stco
func (b *StcoBox) Type() string {
	return "stco"
}

// Size returns the size of the box
func (b *StcoBox) Size() uint64 {
	return uint64(8 + len(b.ChunkOffset)*4)
}

// Dump s the box
func (b *StcoBox) Dump() {
	fmt.Println("Chunk byte offsets:")
	for i, o := range b.ChunkOffset {
		fmt.Printf(" #%d : starts at %d\n", i, o)
	}
}

// Encode encode to the provided writer
func (b *StcoBox) Encode(w io.Writer) error {
	err := EncodeHeader(b, w)
	if err != nil {
		return err
	}
	buf := makebuf(b)
	buf[0] = b.Version
	buf[1], buf[2], buf[3] = b.Flags[0], b.Flags[1], b.Flags[2]
	binary.BigEndian.PutUint32(buf[4:], uint32(len(b.ChunkOffset)))
	for i := range b.ChunkOffset {
		binary.BigEndian.PutUint32(buf[8+4*i:], b.ChunkOffset[i])
	}
	_, err = w.Write(buf)
	return err
}
