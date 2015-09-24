package utils

import (
	"bytes"
	"fmt"
	"io"

	"github.com/ironsmile/nedomi/types"
)

type partWriter struct {
	objID      *types.ObjectID
	cz         types.CacheZone
	partSize   uint64
	startPos   uint64
	currentPos uint64
	length     uint64
	objSize    uint64
	buf        []byte
}

// PartWriterFromContentRange creates a io.WriteCloser that statefully writes sequential parts of
// an object to the supplied storage.
func PartWriterFromContentRange(cz types.CacheZone, objID *types.ObjectID, httpContentRange HTTPContentRange) io.WriteCloser {
	return &partWriter{
		objID:      objID,
		cz:         cz,
		partSize:   cz.Storage.PartSize(),
		startPos:   httpContentRange.Start,
		currentPos: httpContentRange.Start,
		length:     httpContentRange.Length,
		objSize:    httpContentRange.ObjSize,
	}
}

//!TODO: remove
func dbg(s string, args ...interface{}) {
	//fmt.Printf(s, args...)
}

// Fuck go...
func umin(l, r uint64) uint64 {
	if l > r {
		return r
	}
	return l
}

func (pw *partWriter) Write(data []byte) (int, error) {
	dbg("## [%s] Write called with len(data)=%d, partsize=%d\n",
		pw.objID.Path(), len(data), pw.partSize)

	//!TODO: unit test, reduce complexity, reduce int type conversions, better use the buffer/slice methods?

	dataLen := uint64(len(data))
	dataPos := uint64(0)
	remainingData := dataLen
	for remainingData > 0 {
		part := uint32((pw.currentPos - uint64(len(pw.buf))) / pw.partSize)
		fromPartStart := pw.currentPos % pw.partSize
		toPartEnd := pw.partSize - fromPartStart
		dbg("## [%s] Writing part %d; fromPartStart=%d, toPartEnd=%d; remainingData=%d\n",
			pw.objID.Path(), part, fromPartStart, toPartEnd, remainingData)

		if pw.buf == nil {
			dbg("## [%s] Buffer is nil\n", pw.objID.Path())
			if fromPartStart != 0 {
				skip := umin(toPartEnd, remainingData)
				dataPos += skip
				pw.currentPos += skip
				dbg("## [%s] Skipping %d bytes to match part boundary\n", pw.objID.Path(), skip)
			} else {
				dbg("## [%s] Create buffer\n", pw.objID.Path())
				pw.buf = make([]byte, 0, pw.partSize)
			}
		} else {
			dbg("## [%s] Buffer is not nil, len=%d\n", pw.objID.Path(), len(pw.buf))
			if uint64(len(pw.buf)) == pw.partSize {
				dbg("## [%s] Part is finished, save it to disk\n", pw.objID.Path())
				if err := pw.flushBuffer(); err != nil {
					return int(dataPos), err
				}
			} else {
				toWrite := umin(toPartEnd, remainingData)
				dbg("## [%s] Write %d bytes to the buffer\n", pw.objID.Path(), toWrite)
				oldBufLen := uint64(len(pw.buf))
				pw.buf = append(pw.buf, data[dataPos:dataPos+toWrite]...)
				if oldBufLen+toWrite != uint64(len(pw.buf)) {
					return int(dataPos), fmt.Errorf("Partial copy. Expected buffer len to be %d but it is %d\n",
						oldBufLen+toWrite, len(pw.buf))
				}
				dataPos += toWrite
				pw.currentPos += toWrite
			}
		}
		remainingData = dataLen - dataPos
	}
	dbg("## [%s] Write finished, written %d bytes from %d, remaining %d\n",
		pw.objID.Path(), dataPos, len(data), remainingData)

	return int(dataPos), nil
}

func (pw *partWriter) flushBuffer() error {
	if pw.currentPos != pw.objSize && uint64(len(pw.buf)) != pw.partSize {
		return nil
	}
	part := uint32((pw.currentPos - uint64(len(pw.buf))) / pw.partSize)
	idx := &types.ObjectIndex{ObjID: pw.objID, Part: part}
	dbg("## [%s] Saving part %s to the storage (len %d)\n", pw.objID.Path(), idx, len(pw.buf))

	if !pw.cz.Algorithm.ShouldKeep(idx) {
		pw.buf = nil
		return nil
	} else if err := pw.cz.Storage.SavePart(idx, bytes.NewBuffer(pw.buf)); err != nil {
		return err
	}
	pw.cz.Algorithm.PromoteObject(idx)
	pw.buf = nil
	return nil
}

func (pw *partWriter) Close() error {
	if pw.currentPos-pw.startPos != pw.length {
		return fmt.Errorf("PartWriter should have saved %d bytes, but was closed when only %d were received",
			pw.length, pw.currentPos-pw.startPos)
	}
	if pw.buf == nil {
		return nil
	}
	return pw.flushBuffer()
}
