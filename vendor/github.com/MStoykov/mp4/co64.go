package mp4

import (
	"encoding/binary"
	"fmt"
	"io"
)

// Co64Box is Chunk Offset Box (co64 - mandatory)
//
// Contained in : Sample Table box (stbl)
//
// Status: decoded
//
// The table contains the offsets (starting at the beginning of the file) for each chunk of data for the current track.
// A chunk contains samples, the table defining the allocation of samples to each chunk is stsc.
type Co64Box struct {
	Version     byte
	Flags       [3]byte
	ChunkOffset []uint64
}

// DecodeCo64 does what it says on the tin
func DecodeCo64(r io.Reader, size uint64) (Box, error) {
	data, err := read(r, size)
	if err != nil {
		return nil, err
	}

	c := binary.BigEndian.Uint32(data[4:8])
	b := &Co64Box{
		Flags:       [3]byte{data[1], data[2], data[3]},
		Version:     data[0],
		ChunkOffset: make([]uint64, c),
	}
	for i := 0; i < int(c); i++ {
		b.ChunkOffset[i] = binary.BigEndian.Uint64(data[(8 + 8*i):(16 + 8*i)])
	}
	return b, nil
}

// Type returns co64
func (b *Co64Box) Type() string {
	return "co64"
}

// Size returns the size of the box
func (b *Co64Box) Size() uint64 {
	return uint64(8 + len(b.ChunkOffset)*8)
}

// Dump s the box
func (b *Co64Box) Dump() {
	fmt.Println("Chunk byte offsets:")
	for i, o := range b.ChunkOffset {
		fmt.Printf(" #%d : starts at %d\n", i, o)
	}
}

// Encode encode to the provided writer
func (b *Co64Box) Encode(w io.Writer) error {
	if err := EncodeHeader(b, w); err != nil {
		return err
	}
	buf := makebuf(b)
	buf[0] = b.Version
	buf[1], buf[2], buf[3] = b.Flags[0], b.Flags[1], b.Flags[2]
	binary.BigEndian.PutUint32(buf[4:], uint32(len(b.ChunkOffset)))
	for i := range b.ChunkOffset {
		binary.BigEndian.PutUint64(buf[8+8*i:], b.ChunkOffset[i])
	}
	_, err := w.Write(buf)
	return err
}
