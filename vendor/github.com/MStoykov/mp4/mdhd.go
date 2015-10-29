package mp4

import (
	"encoding/binary"
	"fmt"
	"io"
	"time"
)

// MdhdBox - Media Header Box (mdhd - mandatory)
//
// Contained in : Media Box (mdia)
//
// Status : only version 0 is decoded. version 1 is not supported
//
// Timescale defines the timescale used for tracks.
// Language is a ISO-639-2/T language code stored as 1bit padding + [3]int5
type MdhdBox struct {
	Version          byte
	Flags            [3]byte
	CreationTime     uint64
	ModificationTime uint64
	Timescale        uint32
	Duration         uint64
	Language         uint16
}

// DecodeMdhd decodes mdhd
func DecodeMdhd(r io.Reader, size uint64) (Box, error) {
	data, err := read(r, size)
	if err != nil {
		return nil, err
	}
	mdhd := &MdhdBox{
		Version: data[0],
		Flags:   [3]byte{data[1], data[2], data[3]},
	}
	if mdhd.Version == 1 {
		mdhd.CreationTime = binary.BigEndian.Uint64(data[4:12])
		mdhd.ModificationTime = binary.BigEndian.Uint64(data[12:20])
		mdhd.Timescale = binary.BigEndian.Uint32(data[20:24])
		mdhd.Duration = binary.BigEndian.Uint64(data[24:32])
		mdhd.Language = binary.BigEndian.Uint16(data[32:34])
	} else {
		mdhd.Version = 0
		mdhd.CreationTime = uint64(binary.BigEndian.Uint32(data[4:8]))
		mdhd.ModificationTime = uint64(binary.BigEndian.Uint32(data[8:12]))
		mdhd.Timescale = binary.BigEndian.Uint32(data[12:16])
		mdhd.Duration = uint64(binary.BigEndian.Uint32(data[16:20]))
		mdhd.Language = binary.BigEndian.Uint16(data[20:22])
	}

	return mdhd, nil
}

// Type returns mdhd
func (b *MdhdBox) Type() string {
	return "mdhd"
}

// Size returns the size of the mdhd
func (b *MdhdBox) Size() uint64 {
	return uint64(24 + uint64(b.Version)*24)
}

// Dump dumps the mdhd
func (b *MdhdBox) Dump() {
	fmt.Printf("Media Header:\n Timescale: %d units/sec\n Duration: %d units (%s)\n", b.Timescale, b.Duration, time.Duration(b.Duration/uint64(b.Timescale))*time.Second)

}

// Encode encodes
func (b *MdhdBox) Encode(w io.Writer) error {
	err := EncodeHeader(b, w)
	if err != nil {
		return err
	}
	buf := makebuf(b)
	buf[0] = b.Version
	buf[1], buf[2], buf[3] = b.Flags[0], b.Flags[1], b.Flags[2]
	if b.Version == 1 {
		binary.BigEndian.PutUint64(buf[4:], b.CreationTime)
		binary.BigEndian.PutUint64(buf[12:], b.ModificationTime)
		binary.BigEndian.PutUint32(buf[20:], b.Timescale)
		binary.BigEndian.PutUint64(buf[24:], b.Duration)
		binary.BigEndian.PutUint16(buf[32:], b.Language)
	} else {
		binary.BigEndian.PutUint32(buf[4:], uint32(b.CreationTime))
		binary.BigEndian.PutUint32(buf[8:], uint32(b.ModificationTime))
		binary.BigEndian.PutUint32(buf[12:], b.Timescale)
		binary.BigEndian.PutUint32(buf[16:], uint32(b.Duration))
		binary.BigEndian.PutUint16(buf[20:], b.Language)
	}
	_, err = w.Write(buf)
	return err
}
