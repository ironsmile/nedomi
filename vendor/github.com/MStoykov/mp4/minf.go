package mp4

import "io"

// Media Information Box (minf - mandatory)
//
// Contained in : Media Box (mdia)
//
// Status: partially decoded (hmhd - hint tracks - and nmhd - null media - are ignored)
type MinfBox struct {
	Vmhd *VmhdBox
	Smhd *SmhdBox
	Stbl *StblBox
	Dinf *DinfBox
	Hdlr *HdlrBox
	Gmhd *GmhdBox
}

func DecodeMinf(r io.Reader, size uint64) (Box, error) {
	l, err := DecodeContainer(r, size)
	if err != nil {
		return nil, err
	}
	m := &MinfBox{}
	for _, b := range l {
		switch b.Type() {
		case "vmhd":
			m.Vmhd = b.(*VmhdBox)
		case "smhd":
			m.Smhd = b.(*SmhdBox)
		case "stbl":
			m.Stbl = b.(*StblBox)
		case "dinf":
			m.Dinf = b.(*DinfBox)
		case "hdlr":
			m.Hdlr = b.(*HdlrBox)
		case "gmhd":
			m.Gmhd = b.(*GmhdBox)
		}
	}
	return m, nil
}

func (b *MinfBox) Type() string {
	return "minf"
}

func (b *MinfBox) Size() uint64 {
	var sz uint64 = 0
	if b.Vmhd != nil {
		sz += AddHeaderSize(b.Vmhd.Size())
	}
	if b.Smhd != nil {
		sz += AddHeaderSize(b.Smhd.Size())
	}
	sz += AddHeaderSize(b.Stbl.Size())
	if b.Dinf != nil {
		sz += AddHeaderSize(b.Dinf.Size())
	}
	if b.Hdlr != nil {
		sz += AddHeaderSize(b.Hdlr.Size())
	}
	if b.Gmhd != nil {
		sz += AddHeaderSize(b.Gmhd.Size())
	}
	return sz
}

func (b *MinfBox) Dump() {
	b.Stbl.Dump()
}

func (b *MinfBox) Encode(w io.Writer) error {
	err := EncodeHeader(b, w)
	if err != nil {
		return err
	}
	if b.Vmhd != nil {
		err = b.Vmhd.Encode(w)
		if err != nil {
			return err
		}
	}
	if b.Smhd != nil {
		err = b.Smhd.Encode(w)
		if err != nil {
			return err
		}
	}
	err = b.Dinf.Encode(w)
	if err != nil {
		return err
	}
	err = b.Stbl.Encode(w)
	if err != nil {
		return err
	}
	if b.Gmhd != nil {
		err = b.Gmhd.Encode(w)
		if err != nil {
			return err
		}
	}
	if b.Hdlr != nil {
		return b.Hdlr.Encode(w)
	}
	return nil
}
