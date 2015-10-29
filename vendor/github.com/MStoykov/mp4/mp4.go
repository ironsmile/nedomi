package mp4

import (
	"io"
	"math"
	"strconv"
	"time"
)

// MP4 -A MPEG-4 media
//
// A MPEG-4 media contains three main boxes :
//
//   ftyp : the file type box
//   moov : the movie box (meta-data)
//   mdat : the media data (chunks and samples)
//
// Other boxes can also be present (pdin, moof, mfra, free, ...), but are not decoded.
type MP4 struct {
	Ftyp  *FtypBox
	Moov  *MoovBox
	Mdat  *MdatBox
	boxes []Box
}

const firstRequestSize = 4096

// Decode decodes a media from a ReadSeeker
func Decode(rr RangeReader) (*MP4, error) {
	v := &MP4{}
	var in, err = rr.RangeRead(0, firstRequestSize)
	if err != nil {
		return nil, err
	}
	var currentOffset uint64
	var leftFromCurrentRequest uint64 = firstRequestSize

	for {
		h, err := DecodeHeader(in)
		if err != nil {
			if err == io.EOF {
				return v, nil
			}
			return nil, err
		}
		if h.Type == "mdat" { // don't decode that now
			v.Mdat = &MdatBox{Offset: currentOffset, ContentSize: RemoveHeaderSize(h.Size)}
			if v.Moov != nil { // done
				return v, nil
			}
			if err := in.Close(); err != nil { // we don't need that
				return nil, err
			}

			currentOffset += h.Size // go after the mdat
			in, err = rr.RangeRead(currentOffset, firstRequestSize)
			if err != nil {
				return nil, err
			}
			leftFromCurrentRequest = firstRequestSize
			continue
		}
		requiredFromRequest := (h.Size + HeaderSizeFor(h.Size))
		if requiredFromRequest > leftFromCurrentRequest {
			var start, length = currentOffset + leftFromCurrentRequest, requiredFromRequest - leftFromCurrentRequest + firstRequestSize
			leftFromCurrentRequest += length
			nextIn, err := rr.RangeRead(start, length)
			if err != nil {
				return nil, err
			}
			in = newMultiReadCloser(in, nextIn)
		}
		box, err := DecodeBox(h, in)
		if err != nil {
			return nil, err
		}
		switch h.Type {
		case "ftyp":
			v.Ftyp = box.(*FtypBox)
		case "moov":
			v.Moov = box.(*MoovBox)
		default:
			v.boxes = append(v.boxes, box)
		}
		leftFromCurrentRequest -= h.Size
		currentOffset += h.Size
	}
}

// Dump displays some information about a media
func (m *MP4) Dump() {
	m.Ftyp.Dump()
	m.Moov.Dump()
}

// Boxes lists the top-level boxes from a media
func (m *MP4) Boxes() []Box {
	return m.boxes
}

// Encode encodes a media to a Writer
func (m *MP4) Encode(w io.Writer) error {
	err := m.Ftyp.Encode(w)
	if err != nil {
		return err
	}
	err = m.Moov.Encode(w)
	if err != nil {
		return err
	}
	for _, b := range m.boxes {
		err = b.Encode(w)
		if err != nil {
			return err
		}
	}
	err = m.Mdat.Encode(w)
	if err != nil {
		return err
	}
	return nil
}

// Size returns the size of the MP4
func (m *MP4) Size() (size uint64) {
	size += AddHeaderSize(m.Ftyp.Size())
	size += AddHeaderSize(m.Moov.Size())
	size += AddHeaderSize(m.Mdat.Size())

	for _, b := range m.Boxes() {
		size += AddHeaderSize(b.Size())
	}

	return
}

// Duration calculates the duration of the media from the mvhd box
func (m *MP4) Duration() time.Duration {
	return time.Second * time.Duration(m.Moov.Mvhd.Duration) / time.Duration(m.Moov.Mvhd.Timescale)
}

// VideoDimensions returns the dimesnions of the first video trak
func (m *MP4) VideoDimensions() (int, int) {
	for _, trak := range m.Moov.Trak {
		h, _ := strconv.ParseFloat(trak.Tkhd.Height.String(), 64)
		w, _ := strconv.ParseFloat(trak.Tkhd.Width.String(), 64)
		if h > 0 && w > 0 {
			return int(math.Floor(w)), int(math.Floor(h))
		}
	}
	return 0, 0
}

// AudioVolume returns the audio volume of the first audio trak
func (m *MP4) AudioVolume() float64 {
	for _, trak := range m.Moov.Trak {
		vol, _ := strconv.ParseFloat(trak.Tkhd.Volume.String(), 64)
		if vol > 0 {
			return vol
		}
	}
	return 0.0
}
