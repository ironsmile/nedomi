package throttle

import (
	"io"
	"math"
	"runtime"
	"time"
)

const (
	minSleep = 10 * time.Millisecond
)

// ThrottledWriter is a writer that throttles what it writes
// it also implements io.ReaderFrom
type ThrottledWriter struct {
	io.Writer
	speed, written int64
	minWrite       int64
	startTime      time.Time
	now            func() time.Time
	sleep          func(time.Duration)
}

// NewThrottleWriter creates a new throttledWriter writing in the provided io.Writer
// speed is the desired speed, while min and max Writes is the minimum or max writes that will be
// made at a time to the underlying writer.
// Notice: if minWrite is bigger than speed it will become equal to it.
func NewThrottleWriter(w io.Writer, speed, minWrite int64) *ThrottledWriter {
	return &ThrottledWriter{
		Writer:   w,
		speed:    speed,
		minWrite: min64(speed, minWrite),
		now:      time.Now,             // TODO: try to cache it each 10 milliseconds ?
		sleep:    sleepWithPooledTimer, // TODO: provide a way to change it
	}
}

func (tw *ThrottledWriter) Write(b []byte) (n int, err error) {
	if tw.startTime.IsZero() {
		tw.startTime = tw.now()
	}
	for nn := 0; n < len(b) && err == nil; n += nn {
		var toWrite = min64(int64(n)+tw.howMuchCanIWrite(), int64(len(b)))
		nn, err = tw.Writer.Write(b[n:toWrite])
		tw.written += int64(nn)
		if err == nil && int64(nn) != (toWrite-int64(n)) {
			err = io.ErrShortWrite
		}
	}
	return
}

// ReadFrom reads from the provided io.Reader while respecting the throttling
func (tw *ThrottledWriter) ReadFrom(r io.Reader) (n int64, err error) {
	if tw.startTime.IsZero() {
		tw.startTime = tw.now()
	}
	var max int64 = math.MaxInt64
	var lr = io.LimitReader(r, 0).(*io.LimitedReader)
	if llr, ok := lr.R.(*io.LimitedReader); ok {
		max = llr.N
		lr.R = llr.R
	}
	for nn := int64(-1); nn != 0 && err == nil; n += nn {
		lr.N = min64(tw.howMuchCanIWrite(), max)
		nn, err = io.Copy(tw.Writer, lr)
		tw.written += int64(nn)
	}
	if err == io.EOF {
		err = nil
	}
	return
}

func (tw *ThrottledWriter) canWriteRightNow() int64 {
	var timePassed = tw.now().Sub(tw.startTime).Nanoseconds() /
		int64(time.Second) // in seconds
	return timePassed*tw.speed - tw.written
}

func (tw *ThrottledWriter) waitAtleastMinWrite() int64 {
	runtime.Gosched()
	var toWriteRightNow = tw.canWriteRightNow()
	if toWriteRightNow >= tw.minWrite {
		return toWriteRightNow
	}
	calculatedSleep := time.Duration((tw.minWrite-toWriteRightNow)/tw.speed) * time.Second
	tw.sleep(maxDur(calculatedSleep, minSleep))
	// canWriteRightNow may return less than minWrite if minWrite is small enough and the speed
	// is big enough but we have slept long enough for atleast a minWrite so if it's less than
	// we write minWrite.
	return max64(tw.canWriteRightNow(), tw.minWrite)
}

func (tw *ThrottledWriter) howMuchCanIWrite() int64 {
	var toWriteRightNow = tw.canWriteRightNow()
	if toWriteRightNow >= tw.minWrite {
		return toWriteRightNow
	}

	return tw.waitAtleastMinWrite()
}
