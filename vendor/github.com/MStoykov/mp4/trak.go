package mp4

import "io"

// Track Box (tkhd - mandatory)
//
// Contained in : Movie Box (moov)
//
// A media file can contain one or more tracks.
type TrakBox struct {
	Tkhd *TkhdBox
	Mdia *MdiaBox
	Edts *EdtsBox
	Udta *UdtaBox
	Tref *TrefBox
}

func DecodeTrak(r io.Reader, size uint64) (Box, error) {
	l, err := DecodeContainer(r, size)
	if err != nil {
		return nil, err
	}
	t := &TrakBox{}
	for _, b := range l {
		switch b.Type() {
		case "tkhd":
			t.Tkhd = b.(*TkhdBox)
		case "mdia":
			t.Mdia = b.(*MdiaBox)
		case "edts":
			t.Edts = b.(*EdtsBox)
		case "udta":
			t.Udta = b.(*UdtaBox)
		case "tref":
			t.Tref = b.(*TrefBox)
		default:
			return nil, &BadFormatErr{
				enclosingBox:  "trak",
				unexpectedBox: b.Type(),
			}
		}
	}
	return t, nil
}

func (b *TrakBox) Type() string {
	return "trak"
}

func (b *TrakBox) Size() uint64 {
	sz := AddHeaderSize(b.Tkhd.Size())
	sz += AddHeaderSize(b.Mdia.Size())
	if b.Edts != nil {
		sz += AddHeaderSize(b.Edts.Size())
	}
	if b.Udta != nil {
		sz += AddHeaderSize(b.Udta.Size())
	}
	if b.Tref != nil {
		sz += AddHeaderSize(b.Tref.Size())
	}
	return sz
}

func (b *TrakBox) Dump() {
	b.Tkhd.Dump()
	if b.Edts != nil {
		b.Edts.Dump()
	}
	b.Mdia.Dump()
}

func (b *TrakBox) Encode(w io.Writer) error {
	err := EncodeHeader(b, w)
	if err != nil {
		return err
	}
	err = b.Tkhd.Encode(w)
	if err != nil {
		return err
	}
	if b.Edts != nil {
		err = b.Edts.Encode(w)
		if err != nil {
			return err
		}
	}
	if b.Udta != nil {
		err = b.Udta.Encode(w)
		if err != nil {
			return err
		}
	}
	if b.Tref != nil {
		err = b.Tref.Encode(w)
		if err != nil {
			return err
		}
	}
	return b.Mdia.Encode(w)
}
