package clip

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/MStoykov/mp4"
)

var (
	// ErrClipOutside is returned when the clip operation is outside the video
	ErrClipOutside = errors.New("clip zone is outside video")
	// ErrInvalidDuration is returned when the provided duration is invalid
	ErrInvalidDuration = errors.New("invalid duration")
)

// ErrorChunkTrunc is custom error type for when a chunk is truncated
type ErrorChunkTrunc struct {
	m      string
	i1, i2 uint64
}

func (e *ErrorChunkTrunc) Error() string {
	return fmt.Sprintf("%s [%d != %d]", e.m, e.i1, e.i2)
}

type chunk struct {
	size      uint64
	oldOffset uint64
	newOffset uint64
}

type trakInfo struct {
	rebuilded bool

	sci          int
	currentChunk int

	index         uint32
	startTC       uint64
	filterBegin   uint64
	currentSample uint32
	firstSample   uint32
}

type clipFilter struct {
	firstChunk   int
	bufferLength int

	offset  uint64
	forskip uint64

	buffer []byte
	chunks []chunk

	m           *mp4.MP4
	rangeReader mp4.RangeReader

	begin time.Duration
}

// Clip is interface for the clip filter
type Clip interface {
	io.WriterTo
	io.Closer
	Size() uint64
}

// New returns a filter that extracts a clip between begin and begin + duration (in seconds, starting at 0)
// Il will try to include a key frame at the beginning, and keeps the same chunks as the origin media
func New(m *mp4.MP4, begin time.Duration, rr mp4.RangeReader) (Clip, error) {
	if begin < 0 {
		return nil, ErrClipOutside
	}

	if begin > m.Duration() {
		return nil, ErrClipOutside
	}

	f := &clipFilter{
		m:           m,
		begin:       begin,
		rangeReader: rr,
	}
	f.buildChunkList()

	bsz := mp4.HeaderSizeFor(f.chunksSize())
	bsz += mp4.AddHeaderSize(f.m.Ftyp.Size())
	bsz += mp4.AddHeaderSize(f.m.Moov.Size())

	for _, b := range f.m.Boxes() {
		bsz += mp4.AddHeaderSize(b.Size())
	}

	// Update chunk offset
	for _, t := range f.m.Moov.Trak {
		if t.Mdia.Minf.Stbl.Stco != nil {
			for i := range t.Mdia.Minf.Stbl.Stco.ChunkOffset {
				t.Mdia.Minf.Stbl.Stco.ChunkOffset[i] += uint32(bsz)
			}
		} else {

			for i := range t.Mdia.Minf.Stbl.Co64.ChunkOffset {
				t.Mdia.Minf.Stbl.Co64.ChunkOffset[i] += bsz
			}
		}
	}

	// Prepare blob with moov and other small atoms
	var Buffer = bytes.NewBuffer(nil)

	if err := f.m.Ftyp.Encode(Buffer); err != nil {
		return nil, err
	}

	if err := f.m.Moov.Encode(Buffer); err != nil {
		return nil, err
	}

	for _, b := range f.m.Boxes() {
		if err := b.Encode(Buffer); err != nil {
			return nil, err
		}
	}

	if err := mp4.EncodeHeader(f.m.Mdat, Buffer); err != nil {
		return nil, err
	}

	f.buffer = Buffer.Bytes()
	f.bufferLength = len(f.buffer)

	f.compactChunks()

	f.m = nil

	return f, nil
}

func (f *clipFilter) Size() uint64 {
	return uint64(f.bufferLength) + f.chunksSize()
}

func (f *clipFilter) Close() error {
	return nil
}

func (f *clipFilter) chunksSize() uint64 {
	var result uint64
	for _, chunk := range f.chunks {
		result += chunk.size
	}
	return result
}

func (f *clipFilter) compactChunks() {
	newChunks := make([]chunk, 0, 4)
	last := f.chunks[0]
	last.newOffset = uint64(f.bufferLength)
	lastBound := last.oldOffset + last.size
	for i := 1; i < len(f.chunks); i++ {
		ch := f.chunks[i]
		if lastBound == ch.oldOffset {
			lastBound += ch.size
			last.size += ch.size
		} else {
			newChunks = append(newChunks, last)
			ch.newOffset = last.newOffset + last.size
			last = ch
			lastBound = ch.oldOffset + ch.size
		}
	}
	newChunks = append(newChunks, last)
	f.chunks = newChunks
}

func (f *clipFilter) WriteTo(w io.Writer) (n int64, err error) {
	var nn int
	var nnn int64
	var readCloser io.ReadCloser
	if nn, err = w.Write(f.buffer); err != nil {
		return
	}

	n += int64(nn)
	for _, c := range f.chunks {
		readCloser, err = f.rangeReader.RangeRead(c.oldOffset, c.size)
		if err != nil {
			return
		}

		nnn, err = io.Copy(w, readCloser)
		n += nnn
		if err != nil {
			err2 := readCloser.Close()
			return n, mp4.AppendError(err, err2)
		}

		if uint64(nnn) != c.size {
			if err == nil {
				err = &ErrorChunkTrunc{"chunk was truncated: WriteTo", c.size, uint64(nnn)}
			}
		}

		err = readCloser.Close()
		if err != nil {
			return
		}
	}

	return
}

func (f *clipFilter) buildChunkList() {
	var sz, mt int
	var mv, off, size, firstTC, lastTC uint64
	var samples, descriptionID, sample, current uint32

	for _, t := range f.m.Moov.Trak {
		if t.Mdia.Minf.Stbl.Stco != nil {
			sz += len(t.Mdia.Minf.Stbl.Stco.ChunkOffset)
		} else {
			sz += len(t.Mdia.Minf.Stbl.Co64.ChunkOffset)
		}
	}

	f.m.Mdat.ContentSize = 0
	f.m.Moov.Mvhd.Duration = 0

	f.chunks = make([]chunk, 0, sz)

	cnt := len(f.m.Moov.Trak)
	ti := make([]trakInfo, cnt, cnt)

	newFirstChunk := make([][]uint32, cnt, cnt)
	newChunkOffset := make([][]uint32, cnt, cnt)
	newChunkOffset64 := make([][]uint64, cnt, cnt)
	newSamplesPerChunk := make([][]uint32, cnt, cnt)
	newSampleDescriptionID := make([][]uint32, cnt, cnt)

	firstChunkSamples := make([]uint32, cnt, cnt)
	firstChunkDescriptionID := make([]uint32, cnt, cnt)

	fbegin := f.begin

	// Find close l-frame fro begin and end
	for tnum, t := range f.m.Moov.Trak {
		var p uint64

		cti := &ti[tnum]

		newFirstChunk[tnum] = make([]uint32, 0, len(t.Mdia.Minf.Stbl.Stsc.FirstChunk))
		if t.Mdia.Minf.Stbl.Stco != nil {
			newChunkOffset[tnum] = make([]uint32, 0, len(t.Mdia.Minf.Stbl.Stco.ChunkOffset))
		} else {
			newChunkOffset64[tnum] = make([]uint64, 0, len(t.Mdia.Minf.Stbl.Co64.ChunkOffset)) // for co64
		}
		newSamplesPerChunk[tnum] = make([]uint32, 0, len(t.Mdia.Minf.Stbl.Stsc.SamplesPerChunk))
		newSampleDescriptionID[tnum] = make([]uint32, 0, len(t.Mdia.Minf.Stbl.Stsc.SampleDescriptionID))

		cti.filterBegin = uint64(int64(fbegin) * int64(t.Mdia.Mdhd.Timescale) / int64(time.Second))

		if stss := t.Mdia.Minf.Stbl.Stss; stss != nil {
			stts := t.Mdia.Minf.Stbl.Stts

			for i := 0; i < len(stss.SampleNumber); i++ {
				tc := uint64(stts.GetTimeCode(stss.SampleNumber[i] - 1))

				if tc > cti.filterBegin {
					cti.filterBegin = p
					fbegin = time.Second * time.Duration(p) / time.Duration(t.Mdia.Mdhd.Timescale)
					break
				}

				p = tc
			}
		}
	}

	// Skip excess chunks
	for tnum, t := range f.m.Moov.Trak {
		cti := &ti[tnum]

		stco := t.Mdia.Minf.Stbl.Stco
		co64 := t.Mdia.Minf.Stbl.Co64
		stsc := t.Mdia.Minf.Stbl.Stsc
		stts := t.Mdia.Minf.Stbl.Stts
		var length int
		if stco != nil {
			length = len(stco.ChunkOffset)
		} else {
			length = len(co64.ChunkOffset)
		}

		for i := 0; length > i; i++ {
			if cti.sci < len(stsc.FirstChunk)-1 && i+1 >= int(stsc.FirstChunk[cti.sci+1]) {
				cti.sci++
			}

			samples = stsc.SamplesPerChunk[cti.sci]

			firstTC = uint64(stts.GetTimeCode(cti.currentSample + 1))
			cti.currentSample += uint32(samples)
			lastTC = uint64(stts.GetTimeCode(cti.currentSample + 1))

			if lastTC < cti.filterBegin {
				continue
			}

			cti.startTC = firstTC
			cti.currentChunk = i
			cti.currentSample -= samples
			cti.firstSample = cti.currentSample + 1

			break
		}

		if cti.currentChunk == length-1 {
			cnt--
			cti.rebuilded = true
		}
	}

	for cnt > 1 {
		mv = 0
		var is32bit bool

		for tnum, t := range f.m.Moov.Trak {
			if ti[tnum].rebuilded {
				continue
			}

			var currentOffest uint64
			if t.Mdia.Minf.Stbl.Stco != nil {
				currentOffest = uint64(t.Mdia.Minf.Stbl.Stco.ChunkOffset[ti[tnum].currentChunk])
				is32bit = true
			} else {
				currentOffest = t.Mdia.Minf.Stbl.Co64.ChunkOffset[ti[tnum].currentChunk]
			}
			if mv == 0 || currentOffest < mv {
				mt = tnum
				mv = currentOffest
			}
		}

		cti := &ti[mt]

		if is32bit {
			newChunkOffset[mt] = append(newChunkOffset[mt], uint32(off))
		} else {
			newChunkOffset64[mt] = append(newChunkOffset64[mt], off)
		}

		stsc := f.m.Moov.Trak[mt].Mdia.Minf.Stbl.Stsc
		stsz := f.m.Moov.Trak[mt].Mdia.Minf.Stbl.Stsz

		if cti.sci < len(stsc.FirstChunk)-1 && cti.currentChunk+1 >= int(stsc.FirstChunk[cti.sci+1]) {
			cti.sci++
		}

		samples := stsc.SamplesPerChunk[cti.sci]
		descriptionID = stsc.SampleDescriptionID[cti.sci]

		size = 0

		for i := 0; i < int(samples); i++ {
			cti.currentSample++
			size += uint64(stsz.GetSampleSize(int(cti.currentSample)))
		}

		off += size
		f.m.Mdat.ContentSize += size

		f.chunks = append(f.chunks, chunk{
			size:      size,
			oldOffset: mv,
		})

		cti.index++

		if samples != firstChunkSamples[mt] || descriptionID != firstChunkDescriptionID[mt] {
			newFirstChunk[mt] = append(newFirstChunk[mt], cti.index)
			newSamplesPerChunk[mt] = append(newSamplesPerChunk[mt], samples)
			newSampleDescriptionID[mt] = append(newSampleDescriptionID[mt], descriptionID)
			firstChunkSamples[mt] = samples
			firstChunkDescriptionID[mt] = descriptionID
		}

		// Go in next chunk
		cti.currentChunk++

		var l int
		if is32bit {
			l = len(f.m.Moov.Trak[mt].Mdia.Minf.Stbl.Stco.ChunkOffset)
		} else {
			l = len(f.m.Moov.Trak[mt].Mdia.Minf.Stbl.Co64.ChunkOffset)

		}
		if cti.currentChunk == l {
			cnt--
			cti.rebuilded = true
		}
	}

	for tnum, t := range f.m.Moov.Trak {
		cti := &ti[tnum]
		stco := t.Mdia.Minf.Stbl.Stco
		co64 := t.Mdia.Minf.Stbl.Co64
		stsc := t.Mdia.Minf.Stbl.Stsc
		stsz := t.Mdia.Minf.Stbl.Stsz
		stts := t.Mdia.Minf.Stbl.Stts

		end := stts.GetTimeCode(cti.currentSample + 1)

		t.Tkhd.Duration = ((uint64(end) - cti.startTC) / uint64(t.Mdia.Mdhd.Timescale)) * uint64(f.m.Moov.Mvhd.Timescale)
		t.Mdia.Mdhd.Duration = uint64(end) - cti.startTC

		if t.Tkhd.Duration > f.m.Moov.Mvhd.Duration {
			f.m.Moov.Mvhd.Duration = t.Tkhd.Duration
		}

		if !cti.rebuilded {
			var addNewChunkOffest func(undex int, off uint64)
			var chunkOffsetLength int
			if stco != nil {
				addNewChunkOffest = func(index int, off uint64) {
					newChunkOffset[index] = append(newChunkOffset[index], uint32(off))
				}

				chunkOffsetLength = len(stco.ChunkOffset)
			} else {
				addNewChunkOffest = func(index int, off uint64) {
					newChunkOffset64[index] = append(newChunkOffset64[index], off)
				}

				chunkOffsetLength = len(co64.ChunkOffset)
			}
			for i := cti.currentChunk; i < chunkOffsetLength; i++ {
				addNewChunkOffest(tnum, off)

				if cti.sci < len(stsc.FirstChunk)-1 && cti.currentChunk+1 >= int(stsc.FirstChunk[cti.sci+1]) {
					cti.sci++
				}

				samples := stsc.SamplesPerChunk[cti.sci]
				descriptionID := stsc.SampleDescriptionID[cti.sci]

				size = 0

				for i := 0; i < int(samples); i++ {
					cti.currentSample++
					size += uint64(stsz.GetSampleSize(int(cti.currentSample)))
				}

				off += size
				f.m.Mdat.ContentSize += size

				newChunk := chunk{size: size}
				if stco != nil {
					newChunk.oldOffset = uint64(stco.ChunkOffset[i])
				} else {
					newChunk.oldOffset = co64.ChunkOffset[i]
				}
				f.chunks = append(f.chunks, newChunk)

				if samples != firstChunkSamples[tnum] || descriptionID != firstChunkDescriptionID[tnum] {
					newFirstChunk[tnum] = append(newFirstChunk[tnum], uint32(i))
					newSamplesPerChunk[tnum] = append(newSamplesPerChunk[tnum], samples)
					newSampleDescriptionID[tnum] = append(newSampleDescriptionID[tnum], descriptionID)
					firstChunkSamples[tnum] = samples
					firstChunkDescriptionID[tnum] = descriptionID
				}

				cti.currentChunk++
			}
		}

		// stts - sample duration
		if stts := t.Mdia.Minf.Stbl.Stts; stts != nil {
			sample = 1
			current = 0

			firstSample := cti.firstSample
			currentSample := cti.currentSample

			oldSampleCount := stts.SampleCount
			oldSampleTimeDelta := stts.SampleTimeDelta

			newSampleCount := make([]uint32, 0, len(oldSampleCount))
			newSampleTimeDelta := make([]uint32, 0, len(oldSampleTimeDelta))

			for i := 0; i < len(oldSampleCount) && sample < currentSample; i++ {
				if sample+oldSampleCount[i] >= firstSample {
					switch {
					case sample <= firstSample && sample+oldSampleCount[i] > currentSample:
						current = currentSample - firstSample + 1
					case sample < firstSample:
						current = oldSampleCount[i] + sample - firstSample
					case sample+oldSampleCount[i] > currentSample:
						current = oldSampleCount[i] + sample - currentSample
					default:
						current = oldSampleCount[i]
					}

					newSampleCount = append(newSampleCount, current)
					newSampleTimeDelta = append(newSampleTimeDelta, oldSampleTimeDelta[i])
				}

				sample += oldSampleCount[i]
			}

			stts.SampleCount = newSampleCount
			stts.SampleTimeDelta = newSampleTimeDelta
		}

		// stss (key frames)
		if stss := t.Mdia.Minf.Stbl.Stss; stss != nil {
			firstSample := cti.firstSample
			currentSample := cti.currentSample

			oldSampleNumber := stss.SampleNumber
			newSampleNumber := make([]uint32, 0, len(oldSampleNumber))

			for _, n := range oldSampleNumber {
				if n >= firstSample && n <= currentSample {
					newSampleNumber = append(newSampleNumber, n-firstSample+1)
				}
			}

			stss.SampleNumber = newSampleNumber
		}

		// stsz (sample sizes)
		if stsz := t.Mdia.Minf.Stbl.Stsz; stsz != nil {
			firstSample := cti.firstSample
			currentSample := cti.currentSample

			oldSampleSize := stsz.SampleSize

			newSampleSize := make([]uint32, 0, len(oldSampleSize))

			for n, sz := range oldSampleSize {
				if uint32(n) >= firstSample-1 && uint32(n) <= currentSample-1 {
					newSampleSize = append(newSampleSize, sz)
				}
			}

			stsz.SampleSize = newSampleSize
		}

		// ctts - time offsets (b-frames)
		if ctts := t.Mdia.Minf.Stbl.Ctts; ctts != nil {
			sample = 1

			firstSample := cti.firstSample
			currentSample := cti.currentSample

			oldSampleCount := ctts.SampleCount
			oldSampleOffset := ctts.SampleOffset

			newSampleCount := make([]uint32, 0, len(oldSampleCount))
			newSampleOffset := make([]uint32, 0, len(oldSampleOffset))

			for i := 0; i < len(oldSampleCount) && sample < currentSample; i++ {
				if sample+oldSampleCount[i] >= firstSample {
					current := oldSampleCount[i]

					if sample+oldSampleCount[i] > firstSample && sample < firstSample {
						current += sample - firstSample
					}

					if sample+oldSampleCount[i] > currentSample {
						current += currentSample - sample - oldSampleCount[i]
					}

					newSampleCount = append(newSampleCount, current)
					newSampleOffset = append(newSampleOffset, oldSampleOffset[i])
				}

				sample += oldSampleCount[i]
			}

			ctts.SampleCount = newSampleCount
			ctts.SampleOffset = newSampleOffset
		}

		if t.Mdia.Minf.Stbl.Stco != nil {
			t.Mdia.Minf.Stbl.Stco.ChunkOffset = newChunkOffset[tnum]
		} else {
			t.Mdia.Minf.Stbl.Co64.ChunkOffset = newChunkOffset64[tnum]
		}

		t.Mdia.Minf.Stbl.Stsc.FirstChunk = newFirstChunk[tnum]
		t.Mdia.Minf.Stbl.Stsc.SamplesPerChunk = newSamplesPerChunk[tnum]
		t.Mdia.Minf.Stbl.Stsc.SampleDescriptionID = newSampleDescriptionID[tnum]
	}
}
