package mp4

import (
	"encoding/binary"
	"fmt"
	"io"
	"time"
)

// MvhdBox - Movie Header Box (mvhd - mandatory)
//
// Contained in : Movie Box (‘moov’)
//
// Status: version 0 is partially decoded. version 1 is not supported
//
// Contains all media information (duration, ...).
//
// Duration is measured in "time units", and timescale defines the number of time units per second.
//
// Only version 0 is decoded.
type MvhdBox struct {
	Version          byte
	Flags            [3]byte
	CreationTime     uint64
	ModificationTime uint64
	Timescale        uint32
	Duration         uint64
	NextTrackID      uint32
	Rate             Fixed32
	Volume           Fixed16
	Matrix           []byte
}

const (
	mvhdSizeVersion0 = 100
	mvhdSizeVersion1 = mvhdSizeVersion0 + 12
)

// DecodeMvhd - decodes
func DecodeMvhd(r io.Reader, size uint64) (Box, error) {
	// !TODO use size
	var data = make([]byte, 4)
	if _, err := r.Read(data); err != nil {
		return nil, err
	}
	mvhd := &MvhdBox{
		Version: data[0],
		Flags:   [3]byte{data[1], data[2], data[3]},
	}

	data = append(data, make([]byte, 96+(int(mvhd.Version)*12))...)
	var offset = 20 + (int(mvhd.Version) * 12)
	if _, err := r.Read(data[4:]); err != nil {
		return nil, err
	}
	if mvhd.Version == 1 {
		mvhd.CreationTime = binary.BigEndian.Uint64(data[4:12])
		mvhd.ModificationTime = binary.BigEndian.Uint64(data[12:20])
		mvhd.Timescale = binary.BigEndian.Uint32(data[20:24])
		mvhd.Duration = binary.BigEndian.Uint64(data[24:32])
	} else {
		mvhd.CreationTime = uint64(binary.BigEndian.Uint32(data[4:8]))
		mvhd.ModificationTime = uint64(binary.BigEndian.Uint32(data[8:12]))
		mvhd.Timescale = binary.BigEndian.Uint32(data[12:16])
		mvhd.Duration = uint64(binary.BigEndian.Uint32(data[16:20]))
	}
	mvhd.Rate = fixed32(data[offset : offset+4])
	mvhd.Volume = fixed16(data[offset+4 : offset+6])
	// bit(16) // 2
	// unsigned int(32)[2] // 8
	mvhd.Matrix = data[offset+16 : offset+52]
	// bit(32)[6] == 0
	mvhd.NextTrackID = binary.BigEndian.Uint32(data[offset+76 : offset+80])

	return mvhd, nil
}

// Type return mvhd
func (b *MvhdBox) Type() string {
	return "mvhd"
}

// Size returns the size of the mvhd
func (b *MvhdBox) Size() uint64 {
	if b.Version == 1 {
		return mvhdSizeVersion1
	}
	return mvhdSizeVersion0
}

// Dump dumps
func (b *MvhdBox) Dump() {
	fmt.Printf("Movie Header:\n Timescale: %d units/sec\n Duration: %d units (%s)\n Rate: %s\n Volume: %s\n", b.Timescale, b.Duration, time.Duration(b.Duration/uint64(b.Timescale))*time.Second, b.Rate, b.Volume)
}

// Encode encodes
func (b *MvhdBox) Encode(w io.Writer) error {
	err := EncodeHeader(b, w)
	if err != nil {
		return err
	}
	buf := makebuf(b)
	buf[0] = b.Version
	buf[1], buf[2], buf[3] = b.Flags[0], b.Flags[1], b.Flags[2]
	var offset int
	if b.Version == 1 {
		binary.BigEndian.PutUint64(buf[4:], b.CreationTime)
		binary.BigEndian.PutUint64(buf[12:], b.ModificationTime)
		binary.BigEndian.PutUint32(buf[20:], b.Timescale)
		binary.BigEndian.PutUint64(buf[24:], b.Duration)
		offset = 24 + 8
	} else {
		binary.BigEndian.PutUint32(buf[4:], uint32(b.CreationTime))
		binary.BigEndian.PutUint32(buf[8:], uint32(b.ModificationTime))
		binary.BigEndian.PutUint32(buf[12:], b.Timescale)
		binary.BigEndian.PutUint32(buf[16:], uint32(b.Duration))
		offset = 16 + 4
	}
	binary.BigEndian.PutUint32(buf[offset:], uint32(b.Rate))
	binary.BigEndian.PutUint16(buf[offset+4:], uint16(b.Volume))
	// bit(16) // 2
	// unsigned int(32)[2] // 8
	copy(buf[offset+16:], b.Matrix)
	// bit(32)[6] == 0
	binary.BigEndian.PutUint32(buf[offset+76:], b.NextTrackID)

	_, err = w.Write(buf)
	return err
}
