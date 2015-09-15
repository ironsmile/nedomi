package storage

import (
	"bytes"
	"fmt"
	"io"

	"github.com/ironsmile/nedomi/types"
	"github.com/ironsmile/nedomi/utils"
)

type nopCloser struct {
	io.Writer
}

func (nopCloser) Close() error { return nil }

// NopCloser returns a WriteCloser with a no-op Close method wrapping
// the provided Writer w.
func NopCloser(w io.Writer) io.WriteCloser {
	return nopCloser{w}
}

type multiWriteCloser struct {
	writers []io.WriteCloser
}

func (t *multiWriteCloser) Write(p []byte) (n int, err error) {
	for _, w := range t.writers {
		n, err = w.Write(p)
		if err != nil {
			return
		}
		if n != len(p) {
			err = io.ErrShortWrite
			return
		}
	}
	return len(p), nil
}

func (t *multiWriteCloser) Close() error {
	errors := []error{}
	for _, w := range t.writers {
		errors = append(errors, w.Close())
	}
	return utils.NewCompositeError(errors...)
}

// MultiWriteCloser creates a writer that duplicates its writes to all the
// provided writers, similar to the Unix tee(1) command.
func MultiWriteCloser(writers ...io.WriteCloser) io.WriteCloser {
	w := make([]io.WriteCloser, len(writers))
	copy(w, writers)
	return &multiWriteCloser{w}
}

type partWriter struct {
	objID                *types.ObjectID
	storage              types.Storage
	partSize, currentPos uint64
	buf                  []byte
}

// PartWriter creates a io.WriteCloser that statefully writes sequential parts of
// an object to the supplied storage.
func PartWriter(storage types.Storage, objID *types.ObjectID, startPos, endPos uint64) io.WriteCloser {
	return &partWriter{
		objID:      objID,
		storage:    storage,
		partSize:   storage.PartSize(),
		currentPos: startPos,
	}
}

//!TODO: remove
func dbg(s string, args ...interface{}) {
	//fmt.Printf(s, args...)
}
func (pw *partWriter) CurrentPos() uint64 {
	return pw.CurrentPos()
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
		part := uint32(pw.currentPos / pw.partSize)
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
				idx := &types.ObjectIndex{ObjID: pw.objID, Part: part - 1}
				if err := pw.storage.SavePart(idx, bytes.NewBuffer(pw.buf)); err != nil {
					return int(dataPos), err
				}
				pw.buf = nil
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

func (pw *partWriter) Close() error {
	//!TODO: handle network interruptions and non-full parts due to range (take endPos into account)
	if pw.buf == nil {
		return nil
	}
	part := uint32(pw.currentPos / pw.partSize)
	idx := &types.ObjectIndex{ObjID: pw.objID, Part: part}

	dbg("## [%s] Closing writer, flusing part %s to storage (len %d)\n",
		pw.objID.Path(), idx, len(pw.buf))

	return pw.storage.SavePart(idx, bytes.NewBuffer(pw.buf))
}
