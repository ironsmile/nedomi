package mp4

import "io"

// Data Information Box (dinf - mandatory)
//
// Contained in : Media Information Box (minf) or Meta Box (meta)
//
// Status : decoded
type DinfBox struct {
	Dref *DrefBox
}

func DecodeDinf(r io.Reader, size uint64) (Box, error) {
	l, err := DecodeContainer(r, size)
	if err != nil {
		return nil, err
	}
	d := &DinfBox{}
	for _, b := range l {
		switch b.Type() {
		case "dref":
			d.Dref = b.(*DrefBox)
		default:
			return nil, &BadFormatErr{
				enclosingBox:  "dinf",
				unexpectedBox: b.Type(),
			}
		}
	}
	return d, nil
}

func (b *DinfBox) Type() string {
	return "dinf"
}

func (b *DinfBox) Size() uint64 {
	return AddHeaderSize(b.Dref.Size())
}

func (b *DinfBox) Encode(w io.Writer) error {
	if err := EncodeHeader(b, w); err != nil {
		return err
	}
	return b.Dref.Encode(w)
}
