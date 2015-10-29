package mp4

import "io"

// StblBox - Sample Table Box (stbl - mandatory)
//
// Contained in : Media Information Box (minf)
//
// Status: partially decoded (anything other than stsd, stts, stsc, stss, stsz, stco, ctts is ignored)
//
// The table contains all information relevant to data samples (times, chunks, sizes, ...)
type StblBox struct {
	Stsd *StsdBox
	Stts *SttsBox
	Stss *StssBox
	Stsc *StscBox
	Stsz *StszBox
	Stco *StcoBox
	Co64 *Co64Box
	Ctts *CttsBox
}

// DecodeStbl decodes stbl
func DecodeStbl(r io.Reader, size uint64) (Box, error) {
	l, err := DecodeContainer(r, size)
	if err != nil {
		return nil, err
	}
	s := &StblBox{}
	for _, b := range l {
		switch b.Type() {
		case "stsd":
			s.Stsd = b.(*StsdBox)
		case "stts":
			s.Stts = b.(*SttsBox)
		case "stsc":
			s.Stsc = b.(*StscBox)
		case "stss":
			s.Stss = b.(*StssBox)
		case "stsz":
			s.Stsz = b.(*StszBox)
		case "stco":
			s.Stco = b.(*StcoBox)
		case "co64":
			s.Co64 = b.(*Co64Box)
		case "ctts":
			s.Ctts = b.(*CttsBox)
		}
	}
	return s, nil
}

// Type returns stbl
func (b *StblBox) Type() string {
	return "stbl"
}

// Size returns size
func (b *StblBox) Size() uint64 {
	sz := AddHeaderSize(b.Stsd.Size())
	if b.Stts != nil {
		sz += AddHeaderSize(b.Stts.Size())
	}
	if b.Stss != nil {
		sz += AddHeaderSize(b.Stss.Size())
	}
	if b.Stsc != nil {
		sz += AddHeaderSize(b.Stsc.Size())
	}
	if b.Stsz != nil {
		sz += AddHeaderSize(b.Stsz.Size())
	}
	if b.Stco != nil {
		sz += AddHeaderSize(b.Stco.Size())
	}
	if b.Co64 != nil {
		sz += AddHeaderSize(b.Co64.Size())
	}
	if b.Ctts != nil {
		sz += AddHeaderSize(b.Ctts.Size())
	}
	return sz
}

// Dump dumps
func (b *StblBox) Dump() {
	if b.Stsc != nil {
		b.Stsc.Dump()
	}
	if b.Stts != nil {
		b.Stts.Dump()
	}
	if b.Stsz != nil {
		b.Stsz.Dump()
	}
	if b.Stss != nil {
		b.Stss.Dump()
	}
	if b.Stco != nil {
		b.Stco.Dump()
	}
	if b.Co64 != nil {
		b.Co64.Dump()
	}
}

// Encode encodes
func (b *StblBox) Encode(w io.Writer) error {
	err := EncodeHeader(b, w)
	if err != nil {
		return err
	}
	err = b.Stsd.Encode(w)
	if err != nil {
		return err
	}
	err = b.Stts.Encode(w)
	if err != nil {
		return err
	}
	if b.Stss != nil {
		err = b.Stss.Encode(w)
		if err != nil {
			return err
		}
	}
	err = b.Stsc.Encode(w)
	if err != nil {
		return err
	}
	err = b.Stsz.Encode(w)
	if err != nil {
		return err
	}
	if b.Stco != nil {
		err = b.Stco.Encode(w)
		if err != nil {
			return err
		}
	} else {
		err = b.Co64.Encode(w)
		if err != nil {
			return err
		}
	}
	if b.Ctts != nil {
		return b.Ctts.Encode(w)
	}
	return nil
}
