package mp4

import "io"

// User Data Box (udta - optional)
//
// Contained in: Movie Box (moov) or Track Box (trak)
type UdtaBox struct {
	Meta *MetaBox
	Name *NameBox
	Chpl *ChplBox
}

func DecodeUdta(r io.Reader, size uint64) (Box, error) {
	l, err := DecodeContainer(r, size)
	if err != nil {
		return nil, err
	}
	u := &UdtaBox{}
	for _, b := range l {
		switch b.Type() {
		case "meta":
			u.Meta = b.(*MetaBox)
		case "name":
			u.Name = b.(*NameBox)
		case "chpl":
			u.Chpl = b.(*ChplBox)
		default:
			return nil, &BadFormatErr{
				enclosingBox:  "udta",
				unexpectedBox: b.Type(),
			}
		}
	}
	return u, nil
}

func (b *UdtaBox) Type() string {
	return "udta"
}

func (b *UdtaBox) Size() uint64 {
	var sz uint64
	if b.Meta != nil {
		sz += AddHeaderSize(b.Meta.Size())
	}
	if b.Name != nil {
		sz += AddHeaderSize(b.Name.Size())
	}
	if b.Chpl != nil {
		sz += AddHeaderSize(b.Chpl.Size())
	}
	return sz
}

func (b *UdtaBox) Encode(w io.Writer) error {
	err := EncodeHeader(b, w)
	if err != nil {
		return err
	}
	if b.Meta != nil {
		err = b.Meta.Encode(w)
		if err != nil {
			return err
		}
	}
	if b.Name != nil {
		err = b.Name.Encode(w)
		if err != nil {
			return err
		}
	}
	if b.Chpl != nil {
		err = b.Chpl.Encode(w)
		if err != nil {
			return err
		}
	}
	return nil
}
