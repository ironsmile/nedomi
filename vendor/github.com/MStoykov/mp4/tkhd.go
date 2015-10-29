package mp4

import (
	"encoding/binary"
	"fmt"
	"io"
)

// TkhdBox - Track Header Box (tkhd - mandatory)
//
// This box describes the track. Duration is measured in time units (according to the time scale
// defined in the movie header box).
//
// Volume (relevant for audio tracks) is a fixed point number (8 bits + 8 bits). Full volume is 1.0.
// Width and Height (relevant for video tracks) are fixed point numbers (16 bits + 16 bits).
// Video pixels are not necessarily square.
type TkhdBox struct {
	Version          byte
	Flags            [3]byte
	CreationTime     uint64
	ModificationTime uint64
	TrackID          uint32
	Duration         uint64
	Layer            uint16
	AlternateGroup   uint16 // should be int16
	Volume           Fixed16
	Matrix           []byte
	Width, Height    Fixed32
}

// DecodeTkhd te
func DecodeTkhd(r io.Reader, size uint64) (Box, error) {
	// !TODO use size to determine version
	data, err := read(r, size)
	if err != nil {
		return nil, err
	}
	tkhd := &TkhdBox{
		Version: data[0],
		Flags:   [3]byte{data[1], data[2], data[3]},
	}
	var offset = 24
	if tkhd.Version == 1 {
		tkhd.CreationTime = binary.BigEndian.Uint64(data[4:12])
		tkhd.ModificationTime = binary.BigEndian.Uint64(data[12:20])
		tkhd.TrackID = binary.BigEndian.Uint32(data[20:24])
		// uint32 reserved
		tkhd.Duration = binary.BigEndian.Uint64(data[32:36])
		offset = 36
	} else {
		tkhd.CreationTime = uint64(binary.BigEndian.Uint32(data[4:8]))
		tkhd.ModificationTime = uint64(binary.BigEndian.Uint32(data[8:12]))
		tkhd.TrackID = binary.BigEndian.Uint32(data[12:16])
		// uint32 reserved
		tkhd.Duration = uint64(binary.BigEndian.Uint32(data[20:24]))

	}

	offset += 8 // 2 * uint32 reserved
	tkhd.Layer = binary.BigEndian.Uint16(data[offset : offset+2])
	tkhd.AlternateGroup = binary.BigEndian.Uint16(data[offset+2 : offset+4])
	tkhd.Volume = fixed16(data[offset+4 : offset+8])
	tkhd.Matrix = data[offset+8 : offset+44]
	tkhd.Width = fixed32(data[offset+44 : offset+48])
	tkhd.Height = fixed32(data[offset+48 : offset+52])

	return tkhd, nil
}

// Type returns tkhd
func (b *TkhdBox) Type() string {
	return "tkhd"
}

// Size returns the size of the box
func (b *TkhdBox) Size() uint64 {
	if b.Version == 1 {
		return 96
	}
	return 84
}

// Encode encodes to the writer
func (b *TkhdBox) Encode(w io.Writer) error {
	err := EncodeHeader(b, w)
	if err != nil {
		return err
	}
	buf := makebuf(b)
	buf[0] = b.Version
	buf[1], buf[2], buf[3] = b.Flags[0], b.Flags[1], b.Flags[2]
	var offset = 24
	if b.Version == 1 {
		binary.BigEndian.PutUint64(buf[4:], b.CreationTime)
		binary.BigEndian.PutUint64(buf[12:], b.ModificationTime)
		binary.BigEndian.PutUint32(buf[20:], b.TrackID)
		// uint32 reserved
		binary.BigEndian.PutUint64(buf[32:], b.Duration)
		offset = 36
	} else {
		binary.BigEndian.PutUint64(buf[4:], b.CreationTime)
		binary.BigEndian.PutUint64(buf[12:], b.ModificationTime)
		binary.BigEndian.PutUint32(buf[20:], b.TrackID)
		// uint32 reserved
		binary.BigEndian.PutUint64(buf[32:], b.Duration)
	}
	offset += 8
	binary.BigEndian.PutUint16(buf[offset:], b.Layer)
	binary.BigEndian.PutUint16(buf[offset+2:], b.AlternateGroup)
	putFixed16(buf[offset+4:], b.Volume)
	copy(buf[offset+8:], b.Matrix)
	putFixed32(buf[offset+44:], b.Width)
	putFixed32(buf[offset+48:], b.Height)
	_, err = w.Write(buf)
	return err
}

// Dump to terminal
func (b *TkhdBox) Dump() {
	fmt.Println("Track Header:")
	fmt.Printf(" Duration: %d units\n WxH: %sx%s\n", b.Duration, b.Width, b.Height)
}
