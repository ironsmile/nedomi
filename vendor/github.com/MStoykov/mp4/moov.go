package mp4

import (
	"fmt"
	"io"
)

// Movie Box (moov - mandatory)
//
// Status: partially decoded (anything other than mvhd, iods, trak or udta is ignored)
//
// Contains all meta-data. To be able to stream a file, the moov box should be placed before the mdat box.
type MoovBox struct {
	Mvhd *MvhdBox
	Iods *IodsBox
	Trak []*TrakBox
	Udta *UdtaBox
}

func DecodeMoov(r io.Reader, size uint64) (Box, error) {
	l, err := DecodeContainer(r, size)
	if err != nil {
		return nil, err
	}
	m := &MoovBox{}
	for _, b := range l {
		switch b.Type() {
		case "mvhd":
			m.Mvhd = b.(*MvhdBox)
		case "iods":
			m.Iods = b.(*IodsBox)
		case "trak":
			m.Trak = append(m.Trak, b.(*TrakBox))
		case "udta":
			m.Udta = b.(*UdtaBox)
		}
	}
	return m, err
}

func (b *MoovBox) Type() string {
	return "moov"
}

func (b *MoovBox) Size() uint64 {
	var sz uint64 = 0
	if b.Mvhd != nil {
		sz = AddHeaderSize(b.Mvhd.Size())
	}
	if b.Iods != nil {
		sz += AddHeaderSize(b.Iods.Size())
	}
	for _, t := range b.Trak {
		sz += AddHeaderSize(t.Size())
	}
	if b.Udta != nil {
		sz += AddHeaderSize(b.Udta.Size())
	}
	return sz
}

func (b *MoovBox) Dump() {
	if b == nil {
		return
	}
	if b.Mvhd != nil {
		b.Mvhd.Dump()
	}
	for i, t := range b.Trak {
		fmt.Println("Track", i)
		t.Dump()
	}
}

func (b *MoovBox) Encode(w io.Writer) error {
	err := EncodeHeader(b, w)
	if err != nil {
		return err
	}
	err = b.Mvhd.Encode(w)
	if err != nil {
		return err
	}
	if b.Iods != nil {
		err = b.Iods.Encode(w)
		if err != nil {
			return err
		}
	}
	for _, t := range b.Trak {
		err = t.Encode(w)
		if err != nil {
			return err
		}
	}
	if b.Udta != nil {
		return b.Udta.Encode(w)
	}
	return nil
}
