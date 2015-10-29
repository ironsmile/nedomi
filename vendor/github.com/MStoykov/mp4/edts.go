package mp4

import "io"

// Edit Box (edts - optional)
//
// Contained in: Track Box ("trak")
//
// Status: decoded
//
// The edit box maps the presentation timeline to the media-time line
type EdtsBox struct {
	Elst *ElstBox
}

func DecodeEdts(r io.Reader, size uint64) (Box, error) {
	l, err := DecodeContainer(r, size)
	if err != nil {
		return nil, err
	}
	e := &EdtsBox{}
	for _, b := range l {
		switch b.Type() {
		case "elst":
			e.Elst = b.(*ElstBox)
		default:
			return nil, &BadFormatErr{
				enclosingBox:  "edts",
				unexpectedBox: b.Type(),
			}
		}
	}
	return e, nil
}

func (b *EdtsBox) Type() string {
	return "edts"
}

func (b *EdtsBox) Size() uint64 {
	return AddHeaderSize(b.Elst.Size())
}

func (b *EdtsBox) Dump() {
	b.Elst.Dump()
}

func (b *EdtsBox) Encode(w io.Writer) error {
	err := EncodeHeader(b, w)
	if err != nil {
		return err
	}
	return b.Elst.Encode(w)
}
