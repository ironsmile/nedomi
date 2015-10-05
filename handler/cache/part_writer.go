package cache

import (
	"bytes"
	"fmt"
	"io"

	"github.com/ironsmile/nedomi/types"
	"github.com/ironsmile/nedomi/utils/httputils"
)

type partWriter struct {
	objID      *types.ObjectID
	cz         *types.CacheZone
	partSize   uint64
	startPos   uint64
	currentPos uint64
	length     uint64
	objSize    uint64
	buf        []byte
}

// PartWriter creates a io.WriteCloser that statefully writes sequential parts of
// an object to the supplied storage.
func PartWriter(cz *types.CacheZone, objID *types.ObjectID, ContentRange httputils.ContentRange) io.WriteCloser {
	return &partWriter{
		objID:      objID,
		cz:         cz,
		partSize:   cz.Storage.PartSize(),
		startPos:   ContentRange.Start,
		currentPos: ContentRange.Start,
		length:     ContentRange.Length,
		objSize:    ContentRange.ObjSize,
	}
}

// Fuck go...
func umin(l, r uint64) uint64 {
	if l > r {
		return r
	}
	return l
}

func (pw *partWriter) Write(data []byte) (int, error) {
	dataLen := uint64(len(data))
	dataPos := uint64(0)
	remainingData := dataLen
	for remainingData > 0 {
		fromPartStart := pw.currentPos % pw.partSize
		toPartEnd := pw.partSize - fromPartStart

		if pw.buf == nil {
			if fromPartStart != 0 {
				skip := umin(toPartEnd, remainingData)
				dataPos += skip
				pw.currentPos += skip
			} else {
				pw.buf = make([]byte, 0, pw.partSize)
			}
		} else {
			if uint64(len(pw.buf)) == pw.partSize {
				if err := pw.flushBuffer(); err != nil {
					return int(dataPos), err
				}
			} else {
				toWrite := umin(toPartEnd, remainingData)
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

	return int(dataPos), nil
}

func (pw *partWriter) flushBuffer() error {
	if pw.currentPos != pw.objSize && uint64(len(pw.buf)) != pw.partSize {
		return nil
	}
	part := uint32((pw.currentPos - uint64(len(pw.buf))) / pw.partSize)
	idx := &types.ObjectIndex{ObjID: pw.objID, Part: part}

	if !pw.cz.Algorithm.ShouldKeep(idx) {
		pw.buf = nil
		return nil
	} else if err := pw.cz.Storage.SavePart(idx, bytes.NewBuffer(pw.buf)); err != nil {
		return err
	}
	pw.buf = nil
	if err := pw.cz.Algorithm.AddObject(idx); err != nil && err != types.ErrAlreadyInCache {
		return err
	}
	pw.cz.Algorithm.PromoteObject(idx)
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
